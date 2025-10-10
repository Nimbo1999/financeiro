import type { HttpClient } from "~/clients/http";
import type { ListReponse } from "~/models/common";
import type { User } from "~/models/user";

interface UsersPaginationParams {
  page: number;
  page_size: number;
  order_by?: string;
  sort?: string;
  full_name?: string;
}

export interface UserService {
  getByEmail(email: string): Promise<User>;
  getList(params?: URLSearchParams): Promise<ListReponse<User>>;
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
}
