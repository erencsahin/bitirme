import dotenv from 'dotenv';
import { z } from 'zod';

if (process.env.NODE_ENV !== 'production') {
  dotenv.config();
}

const configSchema = z.object({
  node: z.object({
    env: z.enum(['development', 'production', 'test']).default('development'),
  }),
  server: z.object({
    port: z.number().default(8001),
    serviceName: z.string().default('user-service'),
  }),
  database: z.object({
    url: z.string(),
  }),
  redis: z.object({
    host: z.string().default('localhost'),
    port: z.number().default(6380),
    password: z.string().optional(),
  }),
  jwt: z.object({
    secret: z.string(),
    expiresIn: z.string().default('24h'),
    refreshSecret: z.string(),
    refreshExpiresIn: z.string().default('7d'),
  }),
  otel: z.object({
    endpoint: z.string().optional(),
    serviceName: z.string().default('user-service'),
  }),
  logging: z.object({
    level: z.string().default('info'),
  }),
  cors: z.object({
    origin: z.string().default('http://localhost:3000'),
  }),
  rateLimit: z.object({
    windowMs: z.number().default(900000), // 15 minutes
    maxRequests: z.number().default(100),
  }),
  bcrypt: z.object({
    rounds: z.number().default(10),
  }),
});

export type Config = z.infer<typeof configSchema>;

function loadConfig(): Config {
  const rawConfig = {
    node: {
      env: process.env.NODE_ENV || 'development',
    },
    server: {
      port: parseInt(process.env.PORT || '8001', 10),
      serviceName: process.env.SERVICE_NAME || 'user-service',
    },
    database: {
      url: process.env.DATABASE_URL || '',
    },
    redis: {
      host: process.env.REDIS_HOST || 'localhost',
      port: parseInt(process.env.REDIS_PORT || '6380', 10),
      password: process.env.REDIS_PASSWORD,
    },
    jwt: {
      secret: process.env.JWT_SECRET || '',
      expiresIn: process.env.JWT_EXPIRES_IN || '24h',
      refreshSecret: process.env.JWT_REFRESH_SECRET || '',
      refreshExpiresIn: process.env.JWT_REFRESH_EXPIRES_IN || '7d',
    },
    otel: {
      endpoint: process.env.OTEL_EXPORTER_OTLP_ENDPOINT,
      serviceName: process.env.OTEL_SERVICE_NAME || 'user-service',
    },
    logging: {
      level: process.env.LOG_LEVEL || 'info',
    },
    cors: {
      origin: process.env.CORS_ORIGIN || 'http://localhost:3000',
    },
    rateLimit: {
      windowMs: parseInt(process.env.RATE_LIMIT_WINDOW_MS || '900000', 10),
      maxRequests: parseInt(process.env.RATE_LIMIT_MAX_REQUESTS || '100', 10),
    },
    bcrypt: {
      rounds: parseInt(process.env.BCRYPT_ROUNDS || '10', 10),
    },
  };

  try {
    return configSchema.parse(rawConfig);
  } catch (error) {
    if (error instanceof z.ZodError) {
      console.error('Configuration validation failed:');
      error.errors.forEach((err) => {
        console.error(`  - ${err.path.join('.')}: ${err.message}`);
      });
      process.exit(1);
    }
    throw error;
  }
}

export const config = loadConfig();