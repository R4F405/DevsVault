export type DevsVaultClientOptions = {
  baseUrl: string;
  token: string;
  fetcher?: typeof fetch;
};

export type ResolvedSecret = {
  path: string;
  version: number;
  value: string;
};

export class DevsVaultClient {
  private readonly baseUrl: string;
  private readonly token: string;
  private readonly fetcher: typeof fetch;

  constructor(options: DevsVaultClientOptions) {
    if (!options.baseUrl || !options.token) {
      throw new Error("baseUrl and token are required");
    }
    this.baseUrl = options.baseUrl.replace(/\/$/, "");
    this.token = options.token;
    this.fetcher = options.fetcher ?? fetch;
  }

  async getSecret(path: string): Promise<string> {
    const resolved = await this.resolve(path);
    return resolved.value;
  }

  async resolve(path: string): Promise<ResolvedSecret> {
    if (!path || path.split("/").length !== 4) {
      throw new Error("path must be workspace/project/environment/name");
    }
    const response = await this.fetcher(`${this.baseUrl}/api/v1/secrets/resolve?path=${encodeURIComponent(path)}`, {
      headers: { Authorization: `Bearer ${this.token}` }
    });
    if (!response.ok) {
      throw new Error(`secret resolve failed with status ${response.status}`);
    }
    return response.json() as Promise<ResolvedSecret>;
  }
}