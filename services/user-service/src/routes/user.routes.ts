import { Router } from 'express';
import { UserController } from '../controllers/user.controller';
import { authenticate, authorize } from '../middleware/auth.middleware';
import { validate } from '../middleware/validation.middleware';
import { UpdateUserSchema, ChangePasswordSchema } from '../models/user.types';

export function createUserRouter(userController: UserController): Router {
  const router = Router();

  // All routes require authentication
  router.use(authenticate);

  // Get current user profile
  router.get('/me', userController.getCurrentUser);

  // Get all users (admin only)
  router.get('/', authorize('ADMIN'), userController.getAllUsers);

  // Get user by ID
  router.get('/:id', userController.getUserById);

  // Update user
  router.put('/:id', validate(UpdateUserSchema), userController.updateUser);

  // Delete user
  router.delete('/:id', userController.deleteUser);

  // Change password
  router.patch('/:id/password', validate(ChangePasswordSchema), userController.changePassword);

  return router;
}