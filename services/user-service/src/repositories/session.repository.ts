import { Session } from '@prisma/client';
import { prisma } from '../config/database';

export class SessionRepository {
  async create(data: { userId: string; token: string; expiresAt: Date }): Promise<Session> {
    return prisma.session.create({
      data,
    });
  }

  async findByToken(token: string): Promise<Session | null> {
    return prisma.session.findUnique({
      where: { token },
      include: { user: true },
    });
  }

  async deleteByToken(token: string): Promise<void> {
    await prisma.session.delete({
      where: { token },
    });
  }

  async deleteByUserId(userId: string): Promise<void> {
    await prisma.session.deleteMany({
      where: { userId },
    });
  }

  async deleteExpired(): Promise<void> {
    await prisma.session.deleteMany({
      where: {
        expiresAt: {
          lt: new Date(),
        },
      },
    });
  }
}