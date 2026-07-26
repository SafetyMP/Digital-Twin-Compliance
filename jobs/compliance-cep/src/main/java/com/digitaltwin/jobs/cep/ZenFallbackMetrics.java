package com.digitaltwin.jobs.cep;

import java.util.concurrent.atomic.AtomicLong;
import java.util.logging.Logger;

/** Counts Decision Service fallbacks so operators can alert on elevated rates. */
public final class ZenFallbackMetrics {
    private static final Logger LOG = Logger.getLogger(ZenFallbackMetrics.class.getName());
    private static final AtomicLong FALLBACKS = new AtomicLong();

    private ZenFallbackMetrics() {}

    public static void record(String ruleCode, Exception cause) {
        long n = FALLBACKS.incrementAndGet();
        LOG.warning("zen_fallback_total=" + n + " rule=" + ruleCode + " error=" + cause.getMessage());
    }

    public static long total() {
        return FALLBACKS.get();
    }
}
