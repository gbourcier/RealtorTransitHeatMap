import { ref, computed, type Ref, type ComputedRef } from "vue";

export interface ListingFiltersState {
    maxPrice: Ref<number | null>;
    maxCommuteSec: Ref<number | null>;
    newWithinDays: Ref<number | null>;
    minBedrooms: Ref<number | null>;
    minBathrooms: Ref<number | null>;
    minInteriorAreaSqft: Ref<number | null>;
    favoritesOnly: Ref<boolean>;
    includeExpired: Ref<boolean>;
    activeFilterCount: ComputedRef<number>;
    clearAll: () => void;
}

export function useListingFilters(): ListingFiltersState {
    const maxPrice = ref<number | null>(null);
    const maxCommuteSec = ref<number | null>(null);
    const newWithinDays = ref<number | null>(null);
    const minBedrooms = ref<number | null>(1);
    const minBathrooms = ref<number | null>(1);
    const minInteriorAreaSqft = ref<number | null>(null);
    const favoritesOnly = ref(false);
    const includeExpired = ref(false);

    const activeFilterCount = computed(() => {
        let n = 0;
        if (maxPrice.value != null) n++;
        if (maxCommuteSec.value != null) n++;
        if (newWithinDays.value != null) n++;
        if (minBedrooms.value != null && minBedrooms.value > 1) n++;
        if (minBathrooms.value != null && minBathrooms.value > 1) n++;
        if (minInteriorAreaSqft.value != null) n++;
        if (favoritesOnly.value) n++;
        if (includeExpired.value) n++;
        return n;
    });

    function clearAll() {
        maxPrice.value = null;
        maxCommuteSec.value = null;
        newWithinDays.value = null;
        minBedrooms.value = 1;
        minBathrooms.value = 1;
        minInteriorAreaSqft.value = null;
        favoritesOnly.value = false;
        includeExpired.value = false;
    }

    return {
        maxPrice,
        maxCommuteSec,
        newWithinDays,
        minBedrooms,
        minBathrooms,
        minInteriorAreaSqft,
        favoritesOnly,
        includeExpired,
        activeFilterCount,
        clearAll,
    };
}
