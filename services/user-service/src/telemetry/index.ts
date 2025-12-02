import { NodeSDK } from '@opentelemetry/sdk-node';
import { getNodeAutoInstrumentations } from '@opentelemetry/auto-instrumentations-node';
import { OTLPTraceExporter } from '@opentelemetry/exporter-trace-otlp-http';
import { Resource } from '@opentelemetry/resources';
import { SEMRESATTRS_SERVICE_NAME } from '@opentelemetry/semantic-conventions';
import { config } from '../config';
import { logger } from '../utils/logger';

let sdk: NodeSDK | null = null;

export function initTelemetry(): void {
  if (!config.otel.endpoint) {
    logger.warn('OpenTelemetry endpoint not configured, skipping telemetry initialization');
    return;
  }

  try {
    const traceExporter = new OTLPTraceExporter({
      url: `${config.otel.endpoint}/v1/traces`,
    });

    sdk = new NodeSDK({
      resource: new Resource({
        [SEMRESATTRS_SERVICE_NAME]: config.otel.serviceName,
      }),
      traceExporter,
      instrumentations: [
        getNodeAutoInstrumentations({
          '@opentelemetry/instrumentation-fs': {
            enabled: false,
          },
        }),
      ],
    });

    sdk.start();
    logger.info('OpenTelemetry instrumentation initialized');
  } catch (error) {
    logger.error('Failed to initialize OpenTelemetry:', error);
  }
}

export async function shutdownTelemetry(): Promise<void> {
  if (sdk) {
    try {
      await sdk.shutdown();
      logger.info('OpenTelemetry instrumentation shutdown complete');
    } catch (error) {
      logger.error('Error shutting down OpenTelemetry:', error);
    }
  }
}