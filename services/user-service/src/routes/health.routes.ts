import { Router, Request, Response } from 'express';
import { prisma } from '../config/database';
import { getRedisClient } from '../config/redis';
import { ApiResponse } from '../models/api.types';

export function createHealthRouter(): Router {
  const router = Router();

  // Basic health check
  router.get('/health', (_req: Request, res: Response) => {
    const response: ApiResponse = {
      status: 'success',
      data: {
        service: 'user-service',
        status: 'healthy',
        timestamp: new Date().toISOString(),
      },
    };
    res.json(response);
  });

  // Readiness check (with database and redis)
  router.get('/ready', async (_req: Request, res: Response) => {
    try {
      // Check database connection
      await prisma.$queryRaw`SELECT 1`;

      // Check Redis connection
      const redis = getRedisClient();
      await redis.ping();

      const response: ApiResponse = {
        status: 'success',
        data: {
          service: 'user-service',
          status: 'ready',
          database: 'connected',
          redis: 'connected',
          timestamp: new Date().toISOString(),
        },
      };
      res.json(response);
    } catch (error) {
      res.status(503).json({
        status: 'error',
        error: {
          code: 'SERVICE_UNAVAILABLE',
          message: 'Service is not ready',
        },
      } as ApiResponse);
    }
  });

  return router;
}