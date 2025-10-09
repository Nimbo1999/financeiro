import type { User } from "~/models/user";

export interface UserService {
  getByEmail(email: string): Promise<User>;
}
