import { Controller } from '@nestjs/common';
import { MessagePattern } from '@nestjs/microservices';
import { UserService } from './user.service.js';
import { UpdateUserDto } from './dto/update-user.dto.js';

@Controller('users')
export class UserController {
  constructor(private readonly userService: UserService) { }

  @MessagePattern('user.get')
  findById(data: { id: string }) {
    return this.userService.findById(data.id);
  }

  @MessagePattern('user.update')
  update(data: { id: string; userId: string; dto: UpdateUserDto }) {
    return this.userService.update(data.id, data.userId, data.dto);
  }
}
