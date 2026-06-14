package com.digitaltwin.jobs.cep;

import redis.clients.jedis.Jedis;
import redis.clients.jedis.JedisPool;
import redis.clients.jedis.JedisPoolConfig;

public class RedisFeatureStore implements AutoCloseable {
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
            int lastVersion = parseIntOrZero(jedis.get(versionKey));
            if (stateVersion > 0 && stateVersion <= lastVersion) {
                return parseDoubleOrZero(jedis.get(aggregateKey));
            }
            double previousNotional = parseDoubleOrZero(jedis.get(lastNotionalKey));
            double delta = exposureDeltaAmount(previousNotional, newNotionalEur, lastVersion, stateVersion);
            double total = jedis.incrByFloat(aggregateKey, delta);
            jedis.set(lastNotionalKey, Double.toString(newNotionalEur));
            if (stateVersion > 0) {
                jedis.set(versionKey, Integer.toString(stateVersion));
            }
            return total;
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

    private static int parseIntOrZero(String raw) {
        if (raw == null || raw.isEmpty()) {
            return 0;
        }
        try {
            return Integer.parseInt(raw);
        } catch (NumberFormatException e) {
            return 0;
        }
    }

    private static double parseDoubleOrZero(String raw) {
        if (raw == null || raw.isEmpty()) {
            return 0.0;
        }
        try {
            return Double.parseDouble(raw);
        } catch (NumberFormatException e) {
            return 0.0;
        }
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
