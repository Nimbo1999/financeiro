import { FetchClient, type HttpClient } from "~/clients/http";
import type {
  CodeResponse,
  LoginResponse,
  RefreshTokenResponse,
} from "~/models/auth";
import { env } from "~/environment";

export interface AuthService {
  requestCode(email: string): Promise<CodeResponse["data"]>;
  login(email: string, code: string): Promise<LoginResponse["data"]>;
  refreshToken(refreshToken: string): Promise<RefreshTokenResponse["data"]>;
}

export class FinanceiroAuthService implements AuthService {
  constructor(
    private readonly httpClient: HttpClient = FetchClient.new(env.API_BASE_URL)
  ) {}

  async requestCode(email: string): Promise<CodeResponse["data"]> {
    const response = await this.httpClient.post<CodeResponse>(
      "/auth/request-code",
      {
        email,
      }
    );
    if (response.success === false) throw new Error(response.message);
    return response.data;
  }

  async login(email: string, code: string): Promise<LoginResponse["data"]> {
    const response = await this.httpClient.post<LoginResponse>(
      "/auth/verify-code",
      {
        email,
        code,
      }
    );
    if (response.success === false) throw new Error(response.message);
    return response.data;
  }

  async refreshToken(
    refreshToken: string
  ): Promise<RefreshTokenResponse["data"]> {
    const response = await this.httpClient.post<RefreshTokenResponse>(
      "/auth/refresh",
      {
        refresh_token: refreshToken,
      }
    );
    if (response.success === false) throw new Error(response.message);
    return response.data;
  }
}
