import jwt from 'jsonwebtoken';
import { config } from '../config';
import { TokenPayload } from '../models/user.types';
import { UserRole } from '@prisma/client';

export class JwtService {
  generateAccessToken(userId: string, email: string, role: UserRole): string {
    const payload: TokenPayload = {
      userId,
      email,
      role,
      type: 'access',
    };

    return jwt.sign(payload, config.jwt.secret, {
      expiresIn: config.jwt.expiresIn,
    } as any);
  }

  generateRefreshToken(userId: string, email: string, role: UserRole): string {
    const payload: TokenPayload = {
      userId,
      email,
      role,
      type: 'refresh',
    };

    return jwt.sign(payload, config.jwt.refreshSecret, {
      expiresIn: config.jwt.refreshExpiresIn,
    } as any);
  }

  verifyAccessToken(token: string): TokenPayload {
    try {
      const decoded = jwt.verify(token, config.jwt.secret) as TokenPayload;
      if (decoded.type !== 'access') {
        throw new Error('Invalid token type');
      }
      return decoded;
    } catch (error) {
      throw new Error('Invalid or expired access token');
    }
  }

  verifyRefreshToken(token: string): TokenPayload {
    try {
      const decoded = jwt.verify(token, config.jwt.refreshSecret) as TokenPayload;
      if (decoded.type !== 'refresh') {
        throw new Error('Invalid token type');
      }
      return decoded;
    } catch (error) {
      throw new Error('Invalid or expired refresh token');
    }
  }

  getTokenExpiration(token: string): Date {
    const decoded = jwt.decode(token) as any;
    if (!decoded || !decoded.exp) {
      throw new Error('Invalid token');
    }
    return new Date(decoded.exp * 1000);
  }
}

export const jwtService = new JwtService();