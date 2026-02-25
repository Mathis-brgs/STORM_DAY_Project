import {
  IsString,
  MinLength,
  MaxLength,
  Matches,
  IsOptional,
} from 'class-validator';

export class UpdateUserDto {
  @IsOptional()
  @IsString()
  @MinLength(3)
  @MaxLength(20)
  @Matches(/^[a-z0-9_-]+$/, {
    message:
      'username ne peut contenir que des lettres minuscules, chiffres, _ et -',
  })
  username?: string;

  @IsOptional()
  @IsString()
  @MinLength(2)
  @MaxLength(50)
  display_name?: string;

  @IsOptional()
  @IsString()
  avatar_url?: string;
}
