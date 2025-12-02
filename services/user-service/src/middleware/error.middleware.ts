import { Request, Response, NextFunction } from 'express';
import { logger } from '../utils/logger';
import { ApiResponse } from '../models/api.types';

export class AppError extends Error {
  statusCode: number;
  code: string;
  isOperational: boolean;

  constructor(message: string, statusCode: number = 500, code: string = 'INTERNAL_ERROR') {
    super(message);
    this.statusCode = statusCode;
    this.code = code;
    this.isOperational = true;

    Error.captureStackTrace(this, this.constructor);
  }
}

export const errorHandler = (
  err: Error | AppError,
  req: Request,
  res: Response,
  _next: NextFunction
): void => {
  // Default error values
  let statusCode = 500;
  let code = 'INTERNAL_ERROR';
  let message = 'An unexpected error occurred';

  // Handle custom AppError
  if (err instanceof AppError) {
    statusCode = err.statusCode;
    code = err.code;
    message = err.message;
  }
  // Handle known errors
  else if (err.message.includes('not found')) {
    statusCode = 404;
    code = 'NOT_FOUND';
    message = err.message;
  } else if (err.message.includes('already exists')) {
    statusCode = 409;
    code = 'CONFLICT';
    message = err.message;
  } else if (err.message.includes('Invalid') || err.message.includes('incorrect')) {
    statusCode = 400;
    code = 'BAD_REQUEST';
    message = err.message;
  } else if (err.message.includes('Unauthorized')) {
    statusCode = 401;
    code = 'UNAUTHORIZED';
    message = err.message;
  } else if (err.message.includes('Forbidden')) {
    statusCode = 403;
    code = 'FORBIDDEN';
    message = err.message;
  }

  // Log error
  logger.error('Error occurred:', {
    statusCode,
    code,
    message: err.message,
    stack: err.stack,
    path: req.path,
    method: req.method,
  });

  // Send error response
  const response: ApiResponse = {
    status: 'error',
    error: {
      code,
      message,
      ...(process.env.NODE_ENV === 'development' && { stack: err.stack }),
    },
  };

  res.status(statusCode).json(response);
};

export const notFoundHandler = (req: Request, res: Response): void => {
  const response: ApiResponse = {
    status: 'error',
    error: {
      code: 'NOT_FOUND',
      message: `Route ${req.method} ${req.path} not found`,
    },
  };

  res.status(404).json(response);
};