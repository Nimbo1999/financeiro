export interface HttpClient {
  get<T>(url: string, options?: RequestInit): Promise<T>;
  post<T, Body = unknown>(
    url: string,
    body?: Body,
    options?: RequestInit
  ): Promise<T>;
  put<T, Body = unknown>(
    url: string,
    body?: Body,
    options?: RequestInit
  ): Promise<T>;
  delete<T>(url: string, options?: RequestInit): Promise<T>;
}

export class FetchClient implements HttpClient {
  static new(baseURL: string, accessToken?: string) {
    return new FetchClient(baseURL, accessToken);
  }

  constructor(
    private readonly baseURL: string,
    private readonly accessToken?: string
  ) {
    this.baseURL = baseURL;
  }

  async get<T>(url: string, options?: RequestInit): Promise<T> {
    const response = await fetch(`${this.baseURL}${url}`, {
      ...options,
      method: "GET",
      headers: this.getHeaders(options),
    });

    return this.handleResponse<T>(response);
  }

  async post<T, Body = unknown>(
    url: string,
    body?: Body,
    options?: RequestInit
  ): Promise<T> {
    const headers = this.getHeaders(options);
    const response = await fetch(`${this.baseURL}${url}`, {
      ...options,
      method: "POST",
      headers,
      body: this.formatBody(body, headers.get("Content-Type")!),
    });

    return this.handleResponse<T>(response);
  }

  async put<T, Body = unknown>(
    url: string,
    body?: Body,
    options?: RequestInit
  ): Promise<T> {
    const headers = this.getHeaders(options);
    const response = await fetch(`${this.baseURL}${url}`, {
      ...options,
      method: "PUT",
      headers,
      body: this.formatBody(body, headers.get("Content-Type")!),
    });

    return this.handleResponse<T>(response);
  }

  async delete<T>(url: string, options?: RequestInit): Promise<T> {
    const response = await fetch(`${this.baseURL}${url}`, {
      ...options,
      method: "DELETE",
      headers: this.getHeaders(options),
    });

    return this.handleResponse<T>(response);
  }

  private getHeaders(options?: RequestInit): Headers {
    const headers = new Headers({
      "Content-Type": "application/json",
      ...options?.headers,
    });
    if (this.accessToken) {
      headers.set("Authorization", `Bearer ${this.accessToken}`);
    }
    return headers;
  }

  private formatBody<Body = unknown>(
    body?: Body,
    contentType?: string
  ): BodyInit | undefined {
    if (!body) return undefined;
    if (contentType === "application/json" && typeof body === "object") {
      return JSON.stringify(body);
    }
    return body as BodyInit;
  }

  private async handleResponse<T>(response: Response): Promise<T> {
    const isJsonResponse =
      response.headers.get("Content-type") === "application/json";
    if (!response.ok) {
      if (isJsonResponse) {
        const error = await response.json();
        throw error;
      }
      throw new Error(`${response.status}: ${response.statusText}`);
    }
    return isJsonResponse ? response.json() : (response.text() as T);
  }
}
