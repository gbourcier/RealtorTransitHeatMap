import { api, type Paginated } from "./client";

export interface Favorite {
  board: number;
  mls: number;
  latitude: number;
  longitude: number;
  address: string;
  currentPrice: number | null;
  commuteSecondsDowntown: number | null;
  firstSeenAt: number;
  favoritedAt: number;
  slug: string;
  buildingType: number;
  bedroomCount: number;
  bathroomCount: number;
  interiorAreaSqft: number;
  isAvailable: boolean;
}

export type FavoriteSortBy =
  | "favorited_date"
  | "listing_posted_date"
  | "price"
  | "commute";

export interface ListFavoritesParams {
  limit?: number;
  offset?: number;
  sortBy?: FavoriteSortBy;
  sortDir?: "asc" | "desc";
}

export interface FavoriteKey {
  board: number;
  mls: number;
}

export async function listFavorites(
  params: ListFavoritesParams = {},
): Promise<Paginated<Favorite>> {
  const { data } = await api.get<Paginated<Favorite>>("/api/favorites", {
    params,
  });
  return data;
}

export async function addFavorite(board: number, mls: number): Promise<void> {
  await api.post("/api/favorites", { board, mls });
}

export async function removeFavorite(board: number, mls: number): Promise<void> {
  await api.delete(`/api/favorites/${board}/${mls}`);
}

export async function removeFavoritesBatch(
  items: FavoriteKey[],
): Promise<number> {
  const { data } = await api.delete<{ deleted: number }>("/api/favorites", {
    data: { items },
  });
  return data.deleted;
}
