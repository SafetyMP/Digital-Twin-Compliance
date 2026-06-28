package com.digitaltwin.jobs.cep;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import org.apache.flink.api.common.eventtime.WatermarkStrategy;
import org.apache.flink.api.common.serialization.SimpleStringSchema;
import org.apache.flink.connector.base.DeliveryGuarantee;
import org.apache.flink.connector.kafka.sink.KafkaRecordSerializationSchema;
import org.apache.flink.connector.kafka.sink.KafkaSink;
import org.apache.flink.connector.kafka.source.KafkaSource;
import org.apache.flink.streaming.api.datastream.DataStream;
import org.apache.flink.streaming.api.environment.StreamExecutionEnvironment;
import org.apache.flink.streaming.api.functions.ProcessFunction;
import org.apache.flink.util.Collector;

import java.util.HashMap;
import java.util.Map;
import java.util.Optional;
import java.util.logging.Logger;

public class ComplianceCepJob {
    private static final ObjectMapper MAPPER = new ObjectMapper();
    private static final Logger LOG = Logger.getLogger(ComplianceCepJob.class.getName());
    private static final int PARSE_PREVIEW_LEN = 240;

    public static void main(String[] args) throws Exception {
        Map<String, String> params = parseArgs(args);
        JobConfig config = new JobConfig(params);

        StreamExecutionEnvironment env = StreamExecutionEnvironment.getExecutionEnvironment();
        env.setParallelism(config.parallelism);

        KafkaSource<String> paymentsSource = KafkaSource.<String>builder()
                .setBootstrapServers(config.kafkaBrokers)
                .setTopics("domain.events.public.payments")
                .setGroupId(config.paymentsGroupId)
                .setStartingOffsets(config.paymentsOffsets())
                .setValueOnlyDeserializer(new SimpleStringSchema())
                .build();

        KafkaSource<String> twinSource = KafkaSource.<String>builder()
                .setBootstrapServers(config.kafkaBrokers)
                .setTopics("twin.state.updated")
                .setGroupId(config.twinGroupId)
                .setStartingOffsets(config.twinOffsets())
                .setValueOnlyDeserializer(new SimpleStringSchema())
                .build();

        DataStream<String> paymentAlerts = env.fromSource(paymentsSource, WatermarkStrategy.noWatermarks(), "payments")
                .process(new PaymentAlertProcess(config));

        DataStream<String> twinAlerts = env.fromSource(twinSource, WatermarkStrategy.noWatermarks(), "twin-state")
                .process(new TwinAlertProcess(config));

        paymentAlerts.sinkTo(buildAlertSink(config)).name("payment-alerts-sink");
        twinAlerts.sinkTo(buildAlertSink(config)).name("twin-alerts-sink");

        env.execute("compliance-cep");
    }

    static KafkaSink<String> buildAlertSink(JobConfig config) {
        return KafkaSink.<String>builder()
                .setBootstrapServers(config.kafkaBrokers)
                .setDeliveryGuarantee(DeliveryGuarantee.AT_LEAST_ONCE)
                .setRecordSerializer(KafkaRecordSerializationSchema.builder()
                        .setTopic("compliance.alerts")
                        .setValueSerializationSchema(new SimpleStringSchema())
                        .build())
                .build();
    }

    static Map<String, String> parseArgs(String[] args) {
        Map<String, String> params = new HashMap<>();
        for (int i = 0; i < args.length - 1; i++) {
            if (args[i].startsWith("--")) {
                params.put(args[i].substring(2), args[i + 1]);
            }
        }
        return params;
    }

    static String previewPayload(String raw) {
        if (raw == null) {
            return "null";
        }
        if (raw.length() <= PARSE_PREVIEW_LEN) {
            return raw;
        }
        return raw.substring(0, PARSE_PREVIEW_LEN) + "...";
    }

    static String toEnvelopeJson(AlertRecord alert) throws Exception {
        ObjectNode payload = MAPPER.createObjectNode();
        payload.put("alertId", alert.alertId());
        payload.put("ruleCode", alert.ruleCode());
        payload.put("regime", alert.regime());
        payload.put("severity", alert.severity());
        payload.put("status", "Open");
        payload.put("personaId", alert.personaId());
        payload.put("personaType", alert.personaType());
        payload.put("summary", alert.summary());
        payload.put("detectedAt", alert.detectedAt());
        ObjectNode details = payload.putObject("details");
        alert.details().forEach(details::put);

        ObjectNode envelope = MAPPER.createObjectNode();
        envelope.put("eventId", java.util.UUID.randomUUID().toString());
        envelope.put("eventType", "ComplianceAlertRaised");
        envelope.put("eventVersion", "1.0");
        envelope.put("source", "flink-compliance-cep");
        envelope.put("correlationId", alert.correlationId());
        envelope.putNull("causationId");
        envelope.put("timestamp", alert.detectedAt());
        envelope.put("idempotencyKey", alert.idempotencyKey());
        envelope.set("payload", payload);

        return MAPPER.writeValueAsString(envelope);
    }

    static class PaymentAlertProcess extends ProcessFunction<String, String> {
        private transient PatternEngine engine;
        private transient RedisFeatureStore redis;

        private final JobConfig config;

        PaymentAlertProcess(JobConfig config) {
            this.config = config;
        }

        private transient DecisionServiceClient decisionClient;

        @Override
        public void open(org.apache.flink.configuration.Configuration parameters) {
            redis = new RedisFeatureStore(config.redisHost, config.redisPort, config.tenantId);
            if (config.usesDecisionService()) {
                decisionClient = new DecisionServiceClient(config.decisionServiceUrl);
                LOG.info("Phase 3b: INT-M001 uses Decision Service (INT-R001) at " + config.decisionServiceUrl);
            }
            engine = new PatternEngine(config, redis, decisionClient);
        }

        @Override
        public void close() {
            if (redis != null) {
                redis.close();
            }
        }

        @Override
        public void processElement(String value, Context ctx, Collector<String> out) throws Exception {
            Optional<JsonParsers.PaymentEvent> payment = JsonParsers.parsePayment(value);
            if (payment.isEmpty()) {
                LOG.warning("dropped unparseable payment event: " + previewPayload(value));
                return;
            }
            Optional<AlertRecord> alert = engine.onPayment(payment.get());
            if (alert.isPresent()) {
                out.collect(toEnvelopeJson(alert.get()));
            }
        }
    }

    static class TwinAlertProcess extends ProcessFunction<String, String> {
        private transient PatternEngine engine;
        private transient RedisFeatureStore redis;
        private final JobConfig config;

        TwinAlertProcess(JobConfig config) {
            this.config = config;
        }

        private transient DecisionServiceClient decisionClient;

        @Override
        public void open(org.apache.flink.configuration.Configuration parameters) {
            redis = new RedisFeatureStore(config.redisHost, config.redisPort, config.tenantId);
            if (config.usesDecisionService()) {
                decisionClient = new DecisionServiceClient(config.decisionServiceUrl);
                LOG.info("Phase 3b: INT-M002/BASEL-M001 use Decision Service at " + config.decisionServiceUrl);
            }
            engine = new PatternEngine(config, redis, decisionClient);
        }

        @Override
        public void close() {
            if (redis != null) {
                redis.close();
            }
        }

        @Override
        public void processElement(String value, Context ctx, Collector<String> out) throws Exception {
            Optional<JsonParsers.TwinStateEvent> twin = JsonParsers.parseTwinState(value);
            if (twin.isEmpty()) {
                LOG.warning("dropped unparseable twin.state.updated event: " + previewPayload(value));
                return;
            }
            Optional<AlertRecord> alert = engine.onTwinState(twin.get());
            if (alert.isPresent()) {
                out.collect(toEnvelopeJson(alert.get()));
            }
        }
    }
}
