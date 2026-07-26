package com.digitaltwin.jobs.cep;

import redis.clients.jedis.Jedis;
import redis.clients.jedis.JedisPool;
import redis.clients.jedis.JedisPoolConfig;

public class RedisFeatureStore implements AutoCloseable {
    private static final String EXPOSURE_LUA =
            "local versionKey = KEYS[1]\n"
                    + "local lastNotionalKey = KEYS[2]\n"
                    + "local aggregateKey = KEYS[3]\n"
                    + "local stateVersion = tonumber(ARGV[1])\n"
                    + "local newNotional = tonumber(ARGV[2])\n"
                    + "local lastVersion = tonumber(redis.call('GET', versionKey) or '0')\n"
                    + "if stateVersion > 0 and stateVersion <= lastVersion then\n"
                    + "  return tonumber(redis.call('GET', aggregateKey) or '0')\n"
                    + "end\n"
                    + "local previous = tonumber(redis.call('GET', lastNotionalKey) or '0')\n"
                    + "local delta = newNotional - previous\n"
                    + "local total = redis.call('INCRBYFLOAT', aggregateKey, delta)\n"
                    + "redis.call('SET', lastNotionalKey, tostring(newNotional))\n"
                    + "if stateVersion > 0 then\n"
                    + "  redis.call('SET', versionKey, tostring(stateVersion))\n"
                    + "end\n"
                    + "return tonumber(total)\n";

    private final JedisPool pool;
    private final String tenantId;

    public RedisFeatureStore(String host, int port, String tenantId) {
        this.pool = new JedisPool(new JedisPoolConfig(), host, port);
        this.tenantId = tenantId;
    }

    public long incrementVelocity(String accountId) {
        String key = "vel:" + tenantId + ":" + accountId + ":1h";
        try (Jedis jedis = pool.getResource()) {
            long count = jedis.incr(key);
            if (count == 1) {
                jedis.expire(key, 3600);
            }
            return count;
        }
    }

    /**
     * Applies notional delta for an instrument persona, skipping stale stateVersion replays.
     * Uses a Lua script so version/last-notional/aggregate updates are atomic.
     */
    public double applyExposureDelta(
            String personaId,
            String institutionId,
            String counterpartyId,
            double newNotionalEur,
            int stateVersion
    ) {
        String aggregateKey = "exp:" + tenantId + ":" + institutionId + ":" + counterpartyId;
        String lastNotionalKey = "exp-last:" + tenantId + ":" + personaId;
        String versionKey = "exp-ver:" + tenantId + ":" + personaId;
        try (Jedis jedis = pool.getResource()) {
            Object result = jedis.eval(
                    EXPOSURE_LUA,
                    3,
                    versionKey,
                    lastNotionalKey,
                    aggregateKey,
                    Integer.toString(stateVersion),
                    Double.toString(newNotionalEur)
            );
            if (result instanceof Double) {
                return (Double) result;
            }
            if (result instanceof Long) {
                return ((Long) result).doubleValue();
            }
            if (result instanceof String) {
                return Double.parseDouble((String) result);
            }
            return 0.0;
        }
    }

    public static double exposureDeltaAmount(
            double previousNotional,
            double newNotionalEur,
            int lastVersion,
            int stateVersion
    ) {
        if (stateVersion > 0 && stateVersion <= lastVersion) {
            return 0.0;
        }
        return newNotionalEur - previousNotional;
    }

    public void setLcr(String institutionId, double lcr) {
        String key = "lcr:" + tenantId + ":" + institutionId;
        try (Jedis jedis = pool.getResource()) {
            jedis.setex(key, 86400, Double.toString(lcr));
        }
    }

    @Override
    public void close() {
        pool.close();
    }
}
