import { ref, computed, type Ref, type ComputedRef } from "vue";
import {
    listListings,
    type Listing,
    type SortBy,
    type SortDir,
} from "../api/listings";
import type { ListingFiltersState } from "./useListingFilters";

export interface SortOption {
    value: SortBy;
    label: string;
    defaultDir: SortDir;
}

export const sortOptions: SortOption[] = [
    { value: "first_seen_at", label: "Newest", defaultDir: "desc" },
    { value: "price", label: "Price", defaultDir: "asc" },
    { value: "commute_time", label: "Commute", defaultDir: "asc" },
];

export interface ListingsState {
    items: Ref<Listing[]>;
    total: Ref<number>;
    loading: Ref<boolean>;
    error: Ref<string | null>;
    sortBy: Ref<SortBy>;
    sortDir: Ref<SortDir>;
    sortOptions: SortOption[];
    hasMore: ComputedRef<boolean>;
    loadInitial: () => Promise<void>;
    loadMore: () => Promise<void>;
    selectSort: (opt: SortOption) => void;
}

export function useListings(filters: ListingFiltersState): ListingsState {
    const items = ref<Listing[]>([]);
    const total = ref(0);
    const loading = ref(false);
    const error = ref<string | null>(null);
    const sortBy = ref<SortBy>("first_seen_at");
    const sortDir = ref<SortDir>("desc");
    const limit = 50;

    let loadGen = 0;
    let inFlight: Promise<void> | null = null;

    const hasMore = computed(() =>
        items.value.length === 0 ? true : items.value.length < total.value,
    );

    function loadMore(gen: number = loadGen): Promise<void> {
        if (gen !== loadGen) return Promise.resolve();
        if (inFlight) return inFlight.then(() => loadMore(gen));
        if (items.value.length > 0 && items.value.length >= total.value) {
            return Promise.resolve();
        }
        loading.value = true;
        error.value = null;
        inFlight = (async () => {
            try {
                const res = await listListings({
                    limit,
                    offset: items.value.length,
                    sortBy: sortBy.value,
                    sortDir: sortDir.value,
                    ...(filters.maxPrice.value != null && { maxPrice: filters.maxPrice.value }),
                    ...(filters.maxCommuteSec.value != null && { maxCommuteSec: filters.maxCommuteSec.value }),
                    ...(filters.newWithinDays.value != null && { newWithinDays: filters.newWithinDays.value }),
                    ...(filters.minBedrooms.value != null && { minBedrooms: filters.minBedrooms.value }),
                    ...(filters.minBathrooms.value != null && { minBathrooms: filters.minBathrooms.value }),
                    ...(filters.minInteriorAreaSqft.value != null && { minInteriorAreaSqft: filters.minInteriorAreaSqft.value }),
                    ...(filters.favoritesOnly.value && { favoritesOnly: true }),
                    ...(filters.includeExpired.value && { includeExpired: true }),
                });
                if (gen !== loadGen) return;
                items.value = [...items.value, ...res.items];
                total.value = res.total;
            } catch (e: any) {
                if (gen === loadGen) {
                    error.value =
                        e?.response?.data?.error ?? e?.message ?? "failed to load listings";
                }
            } finally {
                if (gen === loadGen) loading.value = false;
                inFlight = null;
            }
        })();
        return inFlight;
    }

    async function loadInitial(): Promise<void> {
        const gen = ++loadGen;
        items.value = [];
        total.value = 0;
        await loadMore(gen);
    }

    function selectSort(opt: SortOption) {
        if (sortBy.value === opt.value) {
            sortDir.value = sortDir.value === "asc" ? "desc" : "asc";
        } else {
            sortBy.value = opt.value;
            sortDir.value = opt.defaultDir;
        }
        loadInitial();
    }

    return {
        items,
        total,
        loading,
        error,
        sortBy,
        sortDir,
        sortOptions,
        hasMore,
        loadInitial,
        loadMore: () => loadMore(),
        selectSort,
    };
}
