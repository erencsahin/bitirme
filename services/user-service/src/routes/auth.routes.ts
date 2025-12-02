import { Router } from 'express';
import { AuthController } from '../controllers/auth.controller';
import { validate } from '../middleware/validation.middleware';
import { LoginUserSchema, RegisterUserSchema } from '../models/user.types';

export function createAuthRouter(authController: AuthController): Router {
  const router = Router();

  router.post('/register', validate(RegisterUserSchema), authController.register);
  router.post('/login', validate(LoginUserSchema), authController.login);
  router.post('/logout', authController.logout);
  router.post('/refresh', authController.refreshToken);
  router.post('/validate',authController.validateToken);

  return router;
}