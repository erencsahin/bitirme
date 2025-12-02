import express, { Application } from 'express';
import cors from 'cors';
import helmet from 'helmet';
import rateLimit from 'express-rate-limit';
import { config } from './config';
import { errorHandler, notFoundHandler } from './middleware/error.middleware';
import { requestLogger } from './middleware/logging.middleware';
import { createHealthRouter } from './routes/health.routes';
import { createAuthRouter } from './routes/auth.routes';
import { createUserRouter } from './routes/user.routes';
import { UserRepository } from './repositories/user.repository';
import { SessionRepository } from './repositories/session.repository';
import { UserService } from './services/user.service';
import { AuthService } from './services/auth.service';
import { UserController } from './controllers/user.controller';
import { AuthController } from './controllers/auth.controller';
import { CacheService, getRedisClient } from './config/redis';

export function createApp(): Application {
  const app = express();

  // Security middleware
  app.use(helmet());
  app.use(
    cors({
      origin: config.cors.origin,
      credentials: true,
    })
  );

  // Rate limiting
  const limiter = rateLimit({
    windowMs: config.rateLimit.windowMs,
    max: config.rateLimit.maxRequests,
    message: {
      status: 'error',
      error: {
        code: 'RATE_LIMIT_EXCEEDED',
        message: 'Too many requests, please try again later',
      },
    },
  });
  app.use('/api', limiter);

  // Body parsing middleware
  app.use(express.json());
  app.use(express.urlencoded({ extended: true }));

  // Request logging
  app.use(requestLogger);

  // Initialize dependencies
  const userRepo = new UserRepository();
  const sessionRepo = new SessionRepository();
  const redis = getRedisClient();
  const cacheService = new CacheService(redis);

  // Services
  const userService = new UserService(userRepo, cacheService);
  const authService = new AuthService(userRepo, sessionRepo, userService);

  // Controllers
  const userController = new UserController(userService);
  const authController = new AuthController(authService);

  // Routes
  app.use('/', createHealthRouter());
  app.use('/api/auth', createAuthRouter(authController));
  app.use('/api/users', createUserRouter(userController));

  // Error handling
  app.use(notFoundHandler);
  app.use(errorHandler);

  return app;
}