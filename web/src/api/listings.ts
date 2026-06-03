import { api, type Paginated } from "./client";

export interface Listing {
  board: number;
  mls: number;
  address: string;
  currentPrice: number | null;
  commuteSecondsDowntown: number | null;
  firstSeenAt: number;
  slug: string;
  bedroomCount: number;
  bathroomCount: number;
  interiorAreaSqft: number;
  isAvailable: boolean;
  isFavorite: boolean;
}

export interface PriceHistory {
  observedAt: number;
  price: number;
}

export interface ListingDetail extends Listing {
  status: string;
  priceHistories: PriceHistory[];
}

export type SortBy = "price" | "first_seen_at" | "commute_time";
export type SortDir = "asc" | "desc";

export interface ListListingsParams {
  limit?: number;
  offset?: number;
  sortBy?: SortBy;
  sortDir?: SortDir;
  maxPrice?: number;
  maxCommuteSec?: number;
  newWithinDays?: number;
  minBedrooms?: number;
  minBathrooms?: number;
  minInteriorAreaSqft?: number;
  favoritesOnly?: boolean;
  includeExpired?: boolean;
}

export interface ListingMapPin {
  board: number;
  mls: number;
  latitude: number;
  longitude: number;
  address: string;
  currentPrice: number | null;
  commuteSecondsDowntown: number | null;
  firstSeenAt: number;
  slug: string;
  bedroomCount: number;
  bathroomCount: number;
  interiorAreaSqft: number;
  isAvailable: boolean;
  isFavorite: boolean;
}

export interface MapFilterParams {
  maxPrice?: number;
  maxCommuteSec?: number;
  newWithinDays?: number;
  minBedrooms?: number;
  minBathrooms?: number;
  minInteriorAreaSqft?: number;
  favoritesOnly?: boolean;
  includeExpired?: boolean;
}

export async function listListings(
  params: ListListingsParams = {},
): Promise<Paginated<Listing>> {
  const { data } = await api.get<Paginated<Listing>>("/api/listings", {
    params,
  });
  return data;
}

export async function listListingsForMap(
  params: MapFilterParams = {},
): Promise<ListingMapPin[]> {
  const { data } = await api.get<ListingMapPin[]>("/api/listings/map", {
    params,
  });
  return data;
}

export async function getListing(
  board: number,
  mls: number,
): Promise<ListingDetail> {
  const { data } = await api.get<ListingDetail>(
    `/api/listings/${board}/${mls}`,
  );
  return data;
}
