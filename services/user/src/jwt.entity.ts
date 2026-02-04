import { Entity, PrimaryColumn, Column, BeforeInsert } from 'typeorm';
import { v7 as uuidv7 } from 'uuid';

@Entity('jwt')
export class Jwt {
  @PrimaryColumn({ type: 'uuid' })
  id: string;

  @Column({ type: 'varchar' })
  jwt: string;

  @Column({ type: 'timestamptz' })
  created_at: Date;

  @Column({ type: 'timestamptz' })
  expirated_at: Date;

  @BeforeInsert()
  generateId() {
    if (!this.id) {
      this.id = uuidv7();
    }
  }
}
