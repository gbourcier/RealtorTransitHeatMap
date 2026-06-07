import { ref, computed, type Ref, type ComputedRef } from "vue";
import type { SavedFilterDefinition } from "../api/savedFilters";

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
    toDefinition: () => SavedFilterDefinition;
    applyDefinition: (def: SavedFilterDefinition) => void;
}

export function useListingFilters(): ListingFiltersState {
    const maxPrice = ref<number | null>(null);
    const maxCommuteSec = ref<number | null>(null);
    const newWithinDays = ref<number | null>(null);
    const minBedrooms = ref<number | null>(null);
    const minBathrooms = ref<number | null>(null);
    const minInteriorAreaSqft = ref<number | null>(null);
    const favoritesOnly = ref(false);
    const includeExpired = ref(false);

    const activeFilterCount = computed(() => {
        let n = 0;
        if (maxPrice.value != null) n++;
        if (maxCommuteSec.value != null) n++;
        if (newWithinDays.value != null) n++;
        if (minBedrooms.value != null) n++;
        if (minBathrooms.value != null) n++;
        if (includeExpired.value) n++;
        return n;
    });

    function toDefinition(): SavedFilterDefinition {
        return {
            maxPrice: maxPrice.value,
            maxCommuteSec: maxCommuteSec.value,
            newWithinDays: newWithinDays.value,
            minBedrooms: minBedrooms.value,
            minBathrooms: minBathrooms.value,
            minInteriorAreaSqft: minInteriorAreaSqft.value,
            favoritesOnly: favoritesOnly.value,
            includeExpired: includeExpired.value,
        };
    }

    function applyDefinition(def: SavedFilterDefinition) {
        maxPrice.value = def.maxPrice;
        maxCommuteSec.value = def.maxCommuteSec;
        newWithinDays.value = def.newWithinDays;
        minBedrooms.value = def.minBedrooms;
        minBathrooms.value = def.minBathrooms;
        minInteriorAreaSqft.value = def.minInteriorAreaSqft;
        favoritesOnly.value = def.favoritesOnly;
        includeExpired.value = def.includeExpired;
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
        toDefinition,
        applyDefinition,
    };
}
