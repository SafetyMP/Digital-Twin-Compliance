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

    public double addExposure(String institutionId, String counterpartyId, double notionalEur) {
        String key = "exp:" + tenantId + ":" + institutionId + ":" + counterpartyId;
        try (Jedis jedis = pool.getResource()) {
            double total = jedis.incrByFloat(key, notionalEur);
            return total;
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
