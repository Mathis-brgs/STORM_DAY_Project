import { Entity, PrimaryGeneratedColumn, Column } from 'typeorm';

@Entity('users_group')
export class UsersGroup {
  @PrimaryGeneratedColumn()
  id: number;

  @Column({ type: 'int' })
  user_id: number;

  @Column({ type: 'int' })
  group_id: number;

  @Column({ type: 'varchar' })
  role: string;
}
