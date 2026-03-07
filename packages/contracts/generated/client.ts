export type HealthResponse = {
  status: string;
  service: string;
  timestamp: string;
};

export class RepoMemoryClient {
  constructor(private readonly baseUrl: string) {}

  async getHealth(): Promise<HealthResponse> {
    const response = await fetch(`${this.baseUrl}/health`, {
      method: "GET",
      headers: {
        "Content-Type": "application/json",
      },
      cache: "no-store",
    });

    if (!response.ok) {
      throw new Error(`health request failed: ${response.status}`);
    }

    return (await response.json()) as HealthResponse;
  }
}