import type { HttpClient } from "~/clients/http";
import type { CreateUserForm } from "~/forms/create-user";
import type { ListReponse } from "~/models/common";
import type { User } from "~/models/user";

export interface UserService {
  getByEmail(email: string): Promise<User>;
  getList(params?: URLSearchParams): Promise<ListReponse<User>>;
  create(formData: CreateUserForm): Promise<User>;
}

export class FinanceiroUserService implements UserService {
  constructor(private readonly httpClient: HttpClient) {}

  async getByEmail(email: string): Promise<User> {
    return await this.httpClient.get<User>(
      `/users/${encodeURIComponent(email)}`
    );
  }

  async getList(params?: URLSearchParams): Promise<ListReponse<User>> {
    return await this.httpClient.get<ListReponse<User>>("/users", params);
  }

  async create(formData: CreateUserForm): Promise<User> {
    return await this.httpClient.post<User>("/users", formData);
  }
}
