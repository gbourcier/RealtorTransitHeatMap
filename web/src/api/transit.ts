import { api } from "./client";

export interface TransitStop {
  latitude: number;
  longitude: number;
  commuteSec: number;
}

export async function listTransitStops(): Promise<TransitStop[]> {
  const { data } = await api.get<TransitStop[]>("/api/transit/stops");
  return data;
}
