interface TokenData {
  access_token: string;
  refresh_token: string;
}

interface AuthenticationData {
  user_id: string;
  email: string;
  tokens: TokenData;
  is_new_user: boolean;
  authenticated_at: string;
}

interface CodeData {
  code_id: string;
  expires_at: string;
  message: string;
}

interface RefreshTokenData {
  tokens: TokenData;
}

interface AuthenticationResponse {
  success: boolean;
  message: string;
}

export interface LoginResponse extends AuthenticationResponse {
  data: AuthenticationData;
}

export interface CodeResponse extends AuthenticationResponse {
  data: CodeData;
}

export interface RefreshTokenResponse extends AuthenticationResponse {
  data: RefreshTokenData;
}
