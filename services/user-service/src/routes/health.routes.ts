import { Router, Request, Response } from 'express';
import { prisma } from '../config/database';
import { getRedisClient } from '../config/redis';
import { ApiResponse } from '../models/api.types';

export function createHealthRouter(): Router {
  const router = Router();

  /**
   * @swagger
   * /health:
   *   get:
   *     summary: Basic health check
   *     description: Returns the health status of the service
   *     tags: [Health]
   *     responses:
   *       200:
   *         description: Service is healthy
   *         content:
   *           application/json:
   *             schema:
   *               $ref: '#/components/schemas/HealthResponse'
   */
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

  /**
   * @swagger
   * /ready:
   *   get:
   *     summary: Readiness check
   *     description: Returns the readiness status including database and Redis connections
   *     tags: [Health]
   *     responses:
   *       200:
   *         description: Service is ready
   *         content:
   *           application/json:
   *             schema:
   *               type: object
   *               properties:
   *                 status:
   *                   type: string
   *                   example: success
   *                 data:
   *                   type: object
   *                   properties:
   *                     service:
   *                       type: string
   *                       example: user-service
   *                     status:
   *                       type: string
   *                       example: ready
   *                     database:
   *                       type: string
   *                       example: connected
   *                     redis:
   *                       type: string
   *                       example: connected
   *                     timestamp:
   *                       type: string
   *                       format: date-time
   *       503:
   *         description: Service is not ready
   *         content:
   *           application/json:
   *             schema:
   *               $ref: '#/components/schemas/ErrorResponse'
   */
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