export type ApiHealth = {
  status: string;
};

const defaultApiBaseUrl = "http://localhost:8080";

export function getApiBaseUrl() {
  return process.env.API_BASE_URL ?? defaultApiBaseUrl;
}

export async function getApiHealth(): Promise<ApiHealth | null> {
  try {
    const response = await fetch(`${getApiBaseUrl()}/api/health`, {
      cache: "no-store",
    });

    if (!response.ok) {
      return null;
    }

    return (await response.json()) as ApiHealth;
  } catch {
    return null;
  }
}
