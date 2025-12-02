import { Request, Response, NextFunction } from 'express';
import { jwtService } from '../utils/jwt';
import { ApiResponse } from '../models/api.types';
import { UserRole } from '@prisma/client';

export const authenticate = async (
  req: Request,
  res: Response,
  next: NextFunction
): Promise<void> => {
  try {
    const authHeader = req.headers.authorization;

    if (!authHeader || !authHeader.startsWith('Bearer ')) {
      res.status(401).json({
        status: 'error',
        error: {
          code: 'UNAUTHORIZED',
          message: 'No token provided',
        },
      } as ApiResponse);
      return;
    }

    const token = authHeader.substring(7); // Remove 'Bearer ' prefix

    try {
      const payload = jwtService.verifyAccessToken(token);
      
      // Attach user info to request
      (req as any).user = payload;
      
      next();
    } catch (error) {
      res.status(401).json({
        status: 'error',
        error: {
          code: 'INVALID_TOKEN',
          message: 'Invalid or expired token',
        },
      } as ApiResponse);
    }
  } catch (error) {
    next(error);
  }
};

export const authorize = (...roles: UserRole[]) => {
  return (req: Request, res: Response, next: NextFunction): void => {
    const user = (req as any).user;

    if (!user) {
      res.status(401).json({
        status: 'error',
        error: {
          code: 'UNAUTHORIZED',
          message: 'Authentication required',
        },
      } as ApiResponse);
      return;
    }

    if (!roles.includes(user.role)) {
      res.status(403).json({
        status: 'error',
        error: {
          code: 'FORBIDDEN',
          message: 'Insufficient permissions',
        },
      } as ApiResponse);
      return;
    }

    next();
  };
};