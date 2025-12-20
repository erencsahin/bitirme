import { UserRepository } from '../repositories/user.repository';
import { SessionRepository } from '../repositories/session.repository';
import { passwordService } from '../utils/password';
import { jwtService } from '../utils/jwt';
import { logger } from '../utils/logger';
import { 
  LoginUserDto, 
  RegisterUserDto, 
  AuthResponseDto, 
  UserResponseDto 
} from '../models/user.types';
import { UserService } from './user.service';

export class AuthService {
  private userRepo: UserRepository;
  private sessionRepo: SessionRepository;
  private userService: UserService;

  constructor(
    userRepo: UserRepository,
    sessionRepo: SessionRepository,
    userService: UserService
  ) {
    this.userRepo = userRepo;
    this.sessionRepo = sessionRepo;
    this.userService = userService;
  }

  async register(data: RegisterUserDto): Promise<AuthResponseDto> {
    // Create user using UserService
    const user = await this.userService.createUser(data);

    // Get full user object for token generation
    const fullUser = await this.userRepo.findById(user.id);
    if (!fullUser) {
      throw new Error('User creation failed');
    }

    // Clear existing sessions just in case
    await this.sessionRepo.deleteByUserId(fullUser.id);

    // Generate tokens
    const accessToken = jwtService.generateAccessToken(
      fullUser.id,
      fullUser.email,
      fullUser.role
    );
    const refreshToken = jwtService.generateRefreshToken(
      fullUser.id,
      fullUser.email,
      fullUser.role
    );

    // Save refresh token session
    const expiresAt = jwtService.getTokenExpiration(refreshToken);
    await this.sessionRepo.create({
      userId: fullUser.id,
      token: refreshToken,
      expiresAt,
    });

    logger.info('User registered successfully', { userId: fullUser.id });

    return {
      user,
      accessToken,
      refreshToken,
    };
  }

  async login(data: LoginUserDto): Promise<AuthResponseDto> {
    // Find user by email
    const user = await this.userRepo.findByEmail(data.email);
    if (!user) {
      throw new Error('Invalid email or password');
    }

    // Check if user is active
    if (!user.isActive) {
      throw new Error('Account is deactivated');
    }

    // Verify password
    const isValidPassword = await passwordService.verify(data.password, user.passwordHash);
    if (!isValidPassword) {
      throw new Error('Invalid email or password');
    }

    // Update last login
    await this.userRepo.updateLastLogin(user.id);
    // Clear existing sessions
    await this.sessionRepo.deleteByUserId(user.id);

    // Generate tokens
    const accessToken = jwtService.generateAccessToken(user.id, user.email, user.role);
    const refreshToken = jwtService.generateRefreshToken(user.id, user.email, user.role);

    // Save refresh token session
    const expiresAt = jwtService.getTokenExpiration(refreshToken);
    await this.sessionRepo.create({
      userId: user.id,
      token: refreshToken,
      expiresAt,
    });

    logger.info('User logged in successfully', { userId: user.id });

    // Convert to response DTO
    const userResponse: UserResponseDto = {
      id: user.id,
      email: user.email,
      username: user.username,
      firstName: user.firstName,
      lastName: user.lastName,
      role: user.role,
      isActive: user.isActive,
      emailVerified: user.emailVerified,
      lastLogin: user.lastLogin,
      createdAt: user.createdAt,
      updatedAt: user.updatedAt,
    };

    return {
      user: userResponse,
      accessToken,
      refreshToken,
    };
  }

  async logout(refreshToken: string): Promise<void> {
    try {
      await this.sessionRepo.deleteByToken(refreshToken);
      logger.info('User logged out successfully');
    } catch (error) {
      logger.warn('Logout failed - token might not exist', { error });
    }
  }

  async refreshToken(refreshToken: string): Promise<AuthResponseDto> {
    // Verify refresh token
    let payload;
    try {
      payload = jwtService.verifyRefreshToken(refreshToken);
    } catch (error) {
      throw new Error('Invalid or expired refresh token');
    }

    // Check if session exists
    const session = await this.sessionRepo.findByToken(refreshToken);
    if (!session) {
      throw new Error('Session not found');
    }

    // Get user
    const user = await this.userRepo.findById(payload.userId);
    if (!user || !user.isActive) {
      throw new Error('User not found or inactive');
    }

    // Delete old session
    await this.sessionRepo.deleteByToken(refreshToken);

    // Generate new tokens
    const newAccessToken = jwtService.generateAccessToken(user.id, user.email, user.role);
    const newRefreshToken = jwtService.generateRefreshToken(user.id, user.email, user.role);

    // Save new refresh token session
    const expiresAt = jwtService.getTokenExpiration(newRefreshToken);
    await this.sessionRepo.create({
      userId: user.id,
      token: newRefreshToken,
      expiresAt,
    });

    logger.info('Token refreshed successfully', { userId: user.id });

    // Convert to response DTO
    const userResponse: UserResponseDto = {
      id: user.id,
      email: user.email,
      username: user.username,
      firstName: user.firstName,
      lastName: user.lastName,
      role: user.role,
      isActive: user.isActive,
      emailVerified: user.emailVerified,
      lastLogin: user.lastLogin,
      createdAt: user.createdAt,
      updatedAt: user.updatedAt,
    };

    return {
      user: userResponse,
      accessToken: newAccessToken,
      refreshToken: newRefreshToken,
    };
  }
  async validateToken(token: string): Promise<{ userId: string; email: string; role: string }> {
    try {
      // Verify access token
      const payload = jwtService.verifyAccessToken(token);

      // Check if user still exists and is active
      const user = await this.userRepo.findById(payload.userId);
      if (!user || !user.isActive) {
        throw new Error('User not found or inactive');
      }

      logger.info('Token validated successfully', { userId: payload.userId });

      return {
        userId: payload.userId,
        email: payload.email,
        role: payload.role,
      };
    } catch (error) {
      logger.warn('Token validation failed', { error });
      throw new Error('Invalid token');
    }
  }
}