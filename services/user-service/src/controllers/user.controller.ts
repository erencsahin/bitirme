import { Request, Response, NextFunction } from 'express';
import { UserService } from '../services/user.service';
import { ApiResponse } from '../models/api.types';
import { UpdateUserDto, ChangePasswordDto } from '../models/user.types';
import { UserRole } from '@prisma/client';

export class UserController {
  private userService: UserService;

  constructor(userService: UserService) {
    this.userService = userService;
  }

  getAllUsers = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      const page = parseInt(req.query.page as string) || 1;
      const limit = parseInt(req.query.limit as string) || 10;
      const role = req.query.role as UserRole | undefined;

      const result = await this.userService.getAllUsers({ page, limit, role });

      const response: ApiResponse = {
        status: 'success',
        data: result.items,
        meta: result.meta,
      };

      res.json(response);
    } catch (error) {
      next(error);
    }
  };

  getUserById = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      const { id } = req.params;
      const user = await this.userService.getUserById(id);

      if (!user) {
        res.status(404).json({
          status: 'error',
          error: {
            code: 'USER_NOT_FOUND',
            message: 'User not found',
          },
        });
        return;
      }

      const response: ApiResponse = {
        status: 'success',
        data: user,
      };

      res.json(response);
    } catch (error) {
      next(error);
    }
  };

  getCurrentUser = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      const userId = (req as any).user.userId;
      const user = await this.userService.getUserById(userId);

      if (!user) {
        res.status(404).json({
          status: 'error',
          error: {
            code: 'USER_NOT_FOUND',
            message: 'User not found',
          },
        });
        return;
      }

      const response: ApiResponse = {
        status: 'success',
        data: user,
      };

      res.json(response);
    } catch (error) {
      next(error);
    }
  };

  updateUser = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      const { id } = req.params;
      const data: UpdateUserDto = req.body;

      // Check if user can update this profile
      const currentUserId = (req as any).user.userId;
      const currentUserRole = (req as any).user.role;

      if (currentUserId !== id && currentUserRole !== 'ADMIN') {
        res.status(403).json({
          status: 'error',
          error: {
            code: 'FORBIDDEN',
            message: 'You can only update your own profile',
          },
        });
        return;
      }

      const user = await this.userService.updateUser(id, data);

      const response: ApiResponse = {
        status: 'success',
        data: user,
      };

      res.json(response);
    } catch (error) {
      next(error);
    }
  };

  deleteUser = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      const { id } = req.params;

      // Check if user can delete this profile
      const currentUserId = (req as any).user.userId;
      const currentUserRole = (req as any).user.role;

      if (currentUserId !== id && currentUserRole !== 'ADMIN') {
        res.status(403).json({
          status: 'error',
          error: {
            code: 'FORBIDDEN',
            message: 'You can only delete your own profile',
          },
        });
        return;
      }

      await this.userService.deleteUser(id);

      const response: ApiResponse = {
        status: 'success',
        data: { message: 'User deleted successfully' },
      };

      res.json(response);
    } catch (error) {
      next(error);
    }
  };

  changePassword = async (req: Request, res: Response, next: NextFunction): Promise<void> => {
    try {
      const { id } = req.params;
      const data: ChangePasswordDto = req.body;

      // Check if user can change this password
      const currentUserId = (req as any).user.userId;

      if (currentUserId !== id) {
        res.status(403).json({
          status: 'error',
          error: {
            code: 'FORBIDDEN',
            message: 'You can only change your own password',
          },
        });
        return;
      }

      await this.userService.changePassword(id, data);

      const response: ApiResponse = {
        status: 'success',
        data: { message: 'Password changed successfully' },
      };

      res.json(response);
    } catch (error) {
      next(error);
    }
  };
}