import { Request, Response, NextFunction } from 'express';
import { AuthService } from '../services/auth.service';
import { ApiResponse } from '../models/api.types';
import { LoginUserDto, RegisterUserDto } from '../models/user.types';

export class AuthController {
  private authService: AuthService;

  constructor(authService: AuthService) {
    this.authService = authService;
  }

  register = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      const data: RegisterUserDto = req.body;
      const result = await this.authService.register(data);

      const response: ApiResponse = {
        status: 'success',
        data: result,
      };

      res.status(201).json(response);
    } catch (error) {
      next(error);
    }
  };

  login = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      const data: LoginUserDto = req.body;
      const result = await this.authService.login(data);

      const response: ApiResponse = {
        status: 'success',
        data: result,
      };

      res.json(response);
    } catch (error) {
      next(error);
    }
  };

  logout = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      const refreshToken = req.body.refreshToken;

      if (!refreshToken) {
        res.status(400).json({
          status: 'error',
          error: {
            code: 'MISSING_REFRESH_TOKEN',
            message: 'Refresh token is required',
          },
        });
        return;
      }

      await this.authService.logout(refreshToken);

      const response: ApiResponse = {
        status: 'success',
        data: { message: 'Logged out successfully' },
      };

      res.json(response);
    } catch (error) {
      next(error);
    }
  };

  refreshToken = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      const refreshToken = req.body.refreshToken;

      if (!refreshToken) {
        res.status(400).json({
          status: 'error',
          error: {
            code: 'MISSING_REFRESH_TOKEN',
            message: 'Refresh token is required',
          },
        });
        return;
      }

      const result = await this.authService.refreshToken(refreshToken);

      const response: ApiResponse = {
        status: 'success',
        data: result,
      };

      res.json(response);
    } catch (error) {
      next(error);
    }
  };
  
  validateToken = async (req: Request, res: Response): Promise<void> => {
    try {
      const authHeader = req.headers.authorization;

      if (!authHeader || !authHeader.startsWith('Bearer ')) {
        res.status(401).json({
          status: 'error',
          error: {
            code: 'INVALID_TOKEN',
            message: 'Invalid or missing token',
          },
        });
        return;
      }

      const token = authHeader.substring(7); // Remove 'Bearer '
      const payload = await this.authService.validateToken(token);

      const response: ApiResponse = {
        status: 'success',
        data: {
          valid: true,
          userId: payload.userId,
          email: payload.email,
        },
      };

      res.json(response);
    } catch (error) {
      res.status(401).json({
        status: 'error',
        error: {
          code: 'INVALID_TOKEN',
          message: 'Token validation failed',
        },
      });
    }
  };
}