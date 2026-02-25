import {
  Injectable,
  NotFoundException,
  ForbiddenException,
} from '@nestjs/common';
import { InjectRepository } from '@nestjs/typeorm';
import { Repository } from 'typeorm';
import { User } from '../user.entity.js';
import { UpdateUserDto } from './dto/update-user.dto.js';

@Injectable()
export class UserService {
  constructor(
    @InjectRepository(User)
    private readonly userRepo: Repository<User>,
  ) {}

  async findById(id: string) {
    const user = await this.userRepo.findOne({ where: { id } });
    if (!user) {
      throw new NotFoundException('Utilisateur introuvable');
    }
    return {
      id: user.id,
      username: user.username,
      display_name: user.display_name,
      email: user.email,
      avatar_url: user.avatar_url,
      created_at: user.created_at,
    };
  }

  async update(id: string, requesterId: string, dto: UpdateUserDto) {
    if (id !== requesterId) {
      throw new ForbiddenException(
        'Vous ne pouvez modifier que votre propre profil',
      );
    }

    const user = await this.userRepo.findOne({ where: { id } });
    if (!user) {
      throw new NotFoundException('Utilisateur introuvable');
    }

    if (dto.username !== undefined) {
      user.username = dto.username;
    }
    if (dto.display_name !== undefined) {
      user.display_name = dto.display_name;
    }
    if (dto.avatar_url !== undefined) {
      user.avatar_url = dto.avatar_url;
    }

    await this.userRepo.save(user);

    return {
      id: user.id,
      username: user.username,
      display_name: user.display_name,
      email: user.email,
      avatar_url: user.avatar_url,
      created_at: user.created_at,
    };
  }

  async search(query: string) {
    const users = await this.userRepo
      .createQueryBuilder('user')
      .where('user.username ILIKE :q', { q: `%${query}%` })
      .orWhere('user.display_name ILIKE :q', { q: `%${query}%` })
      .limit(20)
      .getMany();

    return users.map((u) => ({
      id: u.id,
      username: u.username,
      display_name: u.display_name,
      avatar_url: u.avatar_url,
    }));
  }
}
