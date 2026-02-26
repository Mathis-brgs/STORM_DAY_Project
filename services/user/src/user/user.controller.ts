import { Controller } from '@nestjs/common';
import { MessagePattern } from '@nestjs/microservices';
import { UserService } from './user.service.js';
import { UpdateUserDto } from './dto/update-user.dto.js';

@Controller('users')
export class UserController {
  constructor(private readonly userService: UserService) {}

  @MessagePattern('user.get')
  findById(data: { id: string }) {
    return this.userService.findById(data.id);
  }

  @MessagePattern('user.update')
  update(data: { id: string; userId: string; dto: UpdateUserDto }) {
    return this.userService.update(data.id, data.userId, data.dto);
  }

  @MessagePattern('user.search')
  search(data: { query: string }) {
    return this.userService.search(data.query);
  }

  @MessagePattern('user.status')
  setStatus(data: { userId: string; status: 'online' | 'offline' }) {
    return this.userService.setStatus(data.userId, data.status);
  }

  @MessagePattern('user.status.get')
  getStatus(data: { userId: string }) {
    return this.userService.getStatus(data.userId);
  }
}
