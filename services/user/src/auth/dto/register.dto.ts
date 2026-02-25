import {
  IsEmail,
  IsString,
  MinLength,
  MaxLength,
  Matches,
} from 'class-validator';

export class RegisterDto {
  @IsString()
  @MinLength(3)
  @MaxLength(20)
  @Matches(/^[a-z0-9_-]+$/, {
    message:
      'username ne peut contenir que des lettres minuscules, chiffres, _ et -',
  })
  username: string;

  @IsString()
  @MinLength(2)
  @MaxLength(50)
  display_name: string;

  @IsEmail()
  email: string;

  @IsString()
  @MinLength(6)
  password: string;
}
