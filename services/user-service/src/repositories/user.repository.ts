import { User, UserRole, Prisma } from '@prisma/client';
import { prisma } from '../config/database';
import { PaginationParams } from '../models/api.types';

export class UserRepository {
  async create(data: {
    email: string;
    username: string;
    passwordHash: string;
    firstName?: string;
    lastName?: string;
     role?: UserRole;
  }): Promise<User> {
    return prisma.user.create({
      data,
    });
  }

  async findById(id: string): Promise<User | null> {
    return prisma.user.findUnique({
      where: { id },
    });
  }

  async findByEmail(email: string): Promise<User | null> {
    return prisma.user.findUnique({
      where: { email },
    });
  }

  async findByUsername(username: string): Promise<User | null> {
    return prisma.user.findUnique({
      where: { username },
    });
  }

  async findAll(params: PaginationParams & { role?: UserRole }): Promise<{ users: User[]; total: number }> {
    const { page, limit, role } = params;
    const skip = (page - 1) * limit;

    const where: Prisma.UserWhereInput = role ? { role } : {};

    const [users, total] = await Promise.all([
      prisma.user.findMany({
        where,
        skip,
        take: limit,
        orderBy: { createdAt: 'desc' },
      }),
      prisma.user.count({ where }),
    ]);

    return { users, total };
  }

  async update(id: string, data: Partial<User>): Promise<User> {
    return prisma.user.update({
      where: { id },
      data,
    });
  }

  async delete(id: string): Promise<User> {
    return prisma.user.delete({
      where: { id },
    });
  }

  async updateLastLogin(id: string): Promise<void> {
    await prisma.user.update({
      where: { id },
      data: { lastLogin: new Date() },
    });
  }

  async existsByEmail(email: string): Promise<boolean> {
    const count = await prisma.user.count({
      where: { email },
    });
    return count > 0;
  }

  async existsByUsername(username: string): Promise<boolean> {
    const count = await prisma.user.count({
      where: { username },
    });
    return count > 0;
  }
}