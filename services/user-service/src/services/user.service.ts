import { User, UserRole } from '@prisma/client';
import { UserRepository } from '../repositories/user.repository';
import { passwordService } from '../utils/password';
import { CacheService } from '../config/redis';
import { logger } from '../utils/logger';
import { 
  RegisterUserDto, 
  UpdateUserDto, 
  ChangePasswordDto, 
  UserResponseDto 
} from '../models/user.types';
import { PaginationParams, PaginatedResponse } from '../models/api.types';

export class UserService {
  private userRepo: UserRepository;
  private cacheService: CacheService;
  private readonly CACHE_TTL = 300; // 5 minutes

  constructor(userRepo: UserRepository, cacheService: CacheService) {
    this.userRepo = userRepo;
    this.cacheService = cacheService;
  }

  async createUser(data: RegisterUserDto): Promise<UserResponseDto> {
    // Check if email exists
    const emailExists = await this.userRepo.existsByEmail(data.email);
    if (emailExists) {
      throw new Error('Email already exists');
    }

    // Check if username exists
    const usernameExists = await this.userRepo.existsByUsername(data.username);
    if (usernameExists) {
      throw new Error('Username already exists');
    }

    // Hash password
    const passwordHash = await passwordService.hash(data.password);

    // Create user
    const user = await this.userRepo.create({
      email: data.email,
      username: data.username,
      passwordHash,
      firstName: data.firstName,
      lastName: data.lastName,
      role: data.role ?? UserRole.USER,
    });

    logger.info('User created successfully', { userId: user.id, email: user.email });

    return this.toUserResponse(user);
  }

  async getUserById(id: string): Promise<UserResponseDto | null> {
    // Try cache first
    const cacheKey = `user:${id}`;
    const cachedUser = await this.cacheService.get<UserResponseDto>(cacheKey);
    if (cachedUser) {
      logger.debug('User retrieved from cache', { userId: id });
      return cachedUser;
    }

    // Fetch from database
    const user = await this.userRepo.findById(id);
    if (!user) {
      return null;
    }

    const userResponse = this.toUserResponse(user);

    // Cache the result
    await this.cacheService.set(cacheKey, userResponse, this.CACHE_TTL);

    return userResponse;
  }

  async getUserByEmail(email: string): Promise<User | null> {
    return this.userRepo.findByEmail(email);
  }

  async getAllUsers(params: PaginationParams & { role?: UserRole }): Promise<PaginatedResponse<UserResponseDto>> {
    const { users, total } = await this.userRepo.findAll(params);

    return {
      items: users.map(user => this.toUserResponse(user)),
      meta: {
        page: params.page,
        limit: params.limit,
        total,
        totalPages: Math.ceil(total / params.limit),
      },
    };
  }

  async updateUser(id: string, data: UpdateUserDto): Promise<UserResponseDto> {
    // Check if user exists
    const existingUser = await this.userRepo.findById(id);
    if (!existingUser) {
      throw new Error('User not found');
    }

    // Check email uniqueness if changing
    if (data.email && data.email !== existingUser.email) {
      const emailExists = await this.userRepo.existsByEmail(data.email);
      if (emailExists) {
        throw new Error('Email already exists');
      }
    }

    // Check username uniqueness if changing
    if (data.username && data.username !== existingUser.username) {
      const usernameExists = await this.userRepo.existsByUsername(data.username);
      if (usernameExists) {
        throw new Error('Username already exists');
      }
    }

    // Update user
    const updatedUser = await this.userRepo.update(id, data);

    // Invalidate cache
    await this.cacheService.del(`user:${id}`);

    logger.info('User updated successfully', { userId: id });

    return this.toUserResponse(updatedUser);
  }

  async deleteUser(id: string): Promise<void> {
    const user = await this.userRepo.findById(id);
    if (!user) {
      throw new Error('User not found');
    }

    await this.userRepo.delete(id);

    // Invalidate cache
    await this.cacheService.del(`user:${id}`);

    logger.info('User deleted successfully', { userId: id });
  }

  async changePassword(userId: string, data: ChangePasswordDto): Promise<void> {
    const user = await this.userRepo.findById(userId);
    if (!user) {
      throw new Error('User not found');
    }

    // Verify current password
    const isValid = await passwordService.verify(data.currentPassword, user.passwordHash);
    if (!isValid) {
      throw new Error('Current password is incorrect');
    }

    // Hash new password
    const newPasswordHash = await passwordService.hash(data.newPassword);

    // Update password
    await this.userRepo.update(userId, { passwordHash: newPasswordHash });

    logger.info('Password changed successfully', { userId });
  }

  private toUserResponse(user: User): UserResponseDto {
    return {
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
  }
}