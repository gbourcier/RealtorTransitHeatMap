import { ref, computed, type Ref, type ComputedRef } from "vue";
import type { ListingFiltersState } from "./useListingFilters";
import type { SavedFilter, SavedFilterDefinition } from "../api/savedFilters";
import { useSavedFiltersStore } from "../stores/savedFilters";
import { useAuthStore } from "../stores/auth";

export interface SavedViewsState {
    activeId: Ref<string | null>;
    active: ComputedRef<SavedFilter | null>;
    dirty: ComputedRef<boolean>;
    sortedList: ComputedRef<SavedFilter[]>;
    isDefault: (id: string) => boolean;
    applySaved: (preset: SavedFilter) => void;
    selectAll: () => void;
    selectFavourites: () => void;
    resetFilters: () => void;
    toggleDefault: (id: string) => Promise<void>;
    remove: (id: string) => Promise<void>;
    saveAsNew: (name: string) => Promise<void>;
    updateActive: () => Promise<void>;
}

export function useSavedViews(filters: ListingFiltersState): SavedViewsState {
    const store = useSavedFiltersStore();
    const auth = useAuthStore();
    const activeId = ref<string | null>(null);

    const active = computed(
        () => store.list.find((f) => f.id === activeId.value) ?? null,
    );

    const sortedList = computed(() => {
        const def = auth.defaultFilterId;
        return [...store.list].sort((a, b) => {
            const rank = (id: string) => (id === def ? 0 : 1);
            if (rank(a.id) !== rank(b.id)) return rank(a.id) - rank(b.id);
            return a.createdAt - b.createdAt;
        });
    });

    const dirty = computed(() => {
        const a = active.value;
        if (!a) return false;
        return (
            a.maxPrice !== filters.maxPrice.value ||
            a.maxCommuteSec !== filters.maxCommuteSec.value ||
            a.newWithinDays !== filters.newWithinDays.value ||
            a.minBedrooms !== filters.minBedrooms.value ||
            a.minBathrooms !== filters.minBathrooms.value ||
            a.minInteriorAreaSqft !== filters.minInteriorAreaSqft.value ||
            a.includeExpired !== filters.includeExpired.value
        );
    });

    function definition(): SavedFilterDefinition {
        return {
            ...filters.toDefinition(),
            minInteriorAreaSqft: null,
            favoritesOnly: false,
        };
    }

    function isDefault(id: string): boolean {
        return auth.defaultFilterId === id;
    }

    function applySaved(preset: SavedFilter): void {
        filters.applyDefinition(preset);
        filters.favoritesOnly.value = false;
        activeId.value = preset.id;
    }

    function selectAll(): void {
        resetFilters();
        filters.favoritesOnly.value = false;
    }

    function selectFavourites(): void {
        resetFilters();
        filters.favoritesOnly.value = true;
    }

    function resetFilters(): void {
        filters.maxPrice.value = null;
        filters.maxCommuteSec.value = null;
        filters.newWithinDays.value = null;
        filters.minBedrooms.value = null;
        filters.minBathrooms.value = null;
        filters.minInteriorAreaSqft.value = null;
        filters.favoritesOnly.value = false;
        filters.includeExpired.value = false;
        activeId.value = null;
    }

    async function toggleDefault(id: string): Promise<void> {
        await store.setDefault(isDefault(id) ? null : id);
    }

    async function remove(id: string): Promise<void> {
        await store.remove(id);
        if (activeId.value === id) activeId.value = null;
    }

    async function saveAsNew(name: string): Promise<void> {
        const created = await store.create(name, definition());
        filters.favoritesOnly.value = false;
        activeId.value = created.id;
    }

    async function updateActive(): Promise<void> {
        const a = active.value;
        if (!a) return;
        await store.update(a.id, a.name, definition());
    }

    return {
        activeId,
        active,
        dirty,
        sortedList,
        isDefault,
        applySaved,
        selectAll,
        selectFavourites,
        resetFilters,
        toggleDefault,
        remove,
        saveAsNew,
        updateActive,
    };
}
