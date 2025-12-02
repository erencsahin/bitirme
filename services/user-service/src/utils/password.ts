import bcrypt from 'bcrypt';
import { config } from '../config';

export class PasswordService {
  async hash(password: string): Promise<string> {
    return bcrypt.hash(password, config.bcrypt.rounds);
  }

  async verify(password: string, hash: string): Promise<boolean> {
    return bcrypt.compare(password, hash);
  }
}

export const passwordService = new PasswordService();