interface AuthenticationData {
  user_id: string;
  email: string;
  tokens: {
    access_token: string;
    refresh_token: string;
  };
  is_new_user: boolean;
  authenticated_at: string;
}

interface CodeData {
  code_id: string;
  expires_at: string;
  message: string;
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
