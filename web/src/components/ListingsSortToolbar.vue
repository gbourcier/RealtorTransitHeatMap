<script setup lang="ts">
import { computed } from "vue";
import type { SortBy, SortDir } from "../api/listings";
import type { SortOption } from "../composables/useListings";
import { LISTING_FILTER_BUILDING_TYPES } from "../constants/realtor";

interface Props {
    sortBy: SortBy;
    sortDir: SortDir;
    sortOptions: SortOption[];
    buildingTypes: number[];
    mobile?: boolean;
}

const props = defineProps<Props>();

defineEmits<{
    selectSort: [opt: SortOption];
    selectBuildingType: [id: number | null];
}>();

const buildingTypeOptions = LISTING_FILTER_BUILDING_TYPES;
const buildingTypeActive = computed(() => props.buildingTypes.length > 0);
const buildingTypeLabel = computed(() => {
    if (props.buildingTypes.length === 0) return "Type";
    if (props.buildingTypes.length === 1) {
        return buildingTypeOptions.find((opt) => opt.id === props.buildingTypes[0])?.label ?? "Type";
    }
    return "Types";
});
</script>

<template>
    <div class="list-toolbar" :class="{ 'list-toolbar--mobile': mobile }">
        <div class="sort-tabs" role="tablist" aria-label="Sort listings">
            <button
                v-for="opt in sortOptions"
                :key="opt.value"
                type="button"
                role="tab"
                class="sort-tabs__tab"
                :class="{ 'sort-tabs__tab--active': sortBy === opt.value }"
                :aria-selected="sortBy === opt.value"
                @click="$emit('selectSort', opt)"
            >
                <span class="sort-tabs__label">{{ opt.label }}</span>
                <v-icon v-if="sortBy === opt.value" size="14" class="sort-tabs__dir">
                    {{ sortDir === 'asc' ? 'mdi-arrow-up' : 'mdi-arrow-down' }}
                </v-icon>
            </button>
        </div>
        <v-menu location="bottom end" :offset="6" content-class="building-type-menu-surface">
            <template #activator="{ props: menuProps }">
                <button
                    v-bind="menuProps"
                    type="button"
                    class="building-type-pill"
                    :class="{ 'building-type-pill--active': buildingTypeActive }"
                    :aria-pressed="buildingTypeActive"
                >
                    <span class="building-type-pill__label">{{ buildingTypeLabel }}</span>
                    <v-icon size="14" class="building-type-pill__icon">mdi-chevron-down</v-icon>
                </button>
            </template>
            <v-list class="building-type-menu" density="compact" nav>
                <v-list-item
                    title="Any type"
                    :active="buildingTypes.length === 0"
                    @click="$emit('selectBuildingType', null)"
                />
                <v-list-item
                    v-for="opt in buildingTypeOptions"
                    :key="opt.id"
                    :title="opt.label"
                    :active="buildingTypes.length === 1 && buildingTypes[0] === opt.id"
                    @click="$emit('selectBuildingType', opt.id)"
                />
            </v-list>
        </v-menu>
    </div>
</template>

<style scoped>
.list-toolbar {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    padding: 10px 14px;
    border-bottom: 1px solid rgba(var(--v-theme-on-surface), 0.08);
    flex: 0 0 auto;
}

.list-toolbar--mobile {
    padding: 10px 14px;
    border-bottom: 1px solid rgba(var(--v-theme-on-surface), 0.06);
}

.sort-tabs {
    display: inline-flex;
    align-items: stretch;
    flex: 0 0 auto;
    min-width: 0;
    border: 1px solid rgba(var(--v-theme-on-surface), 0.14);
    border-radius: 999px;
    overflow: hidden;
    background-color: transparent;
}

.sort-tabs__tab {
    flex: 0 0 auto;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 4px;
    min-width: 0;
    height: 30px;
    padding: 0 12px;
    background: transparent;
    border: 0;
    border-left: 1px solid rgba(var(--v-theme-on-surface), 0.14);
    border-radius: 0;
    cursor: pointer;
    color: rgba(var(--v-theme-on-surface), 0.7);
    font-size: 0.8125rem;
    font-weight: 500;
    letter-spacing: normal;
    transition: background-color 120ms ease, color 120ms ease;
}

.sort-tabs__tab:first-child {
    border-left: 0;
}

.sort-tabs__tab:hover {
    background-color: rgba(var(--v-theme-on-surface), 0.05);
    color: rgba(var(--v-theme-on-surface), 0.92);
}

.sort-tabs__tab:focus-visible {
    outline: 2px solid rgb(var(--v-theme-primary));
    outline-offset: -2px;
}

.sort-tabs__tab--active {
    background-color: rgba(var(--v-theme-on-surface), 0.08);
    color: rgba(var(--v-theme-on-surface), 0.98);
    font-weight: 600;
}

.sort-tabs__label {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.sort-tabs__dir {
    flex-shrink: 0;
    opacity: 0.9;
}

.building-type-pill {
    flex: 0 0 auto;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 4px;
    min-width: 0;
    height: 30px;
    padding: 0 12px;
    border: 1px solid rgba(var(--v-theme-on-surface), 0.14);
    border-radius: 999px;
    background: transparent;
    color: rgba(var(--v-theme-on-surface), 0.7);
    cursor: pointer;
    font-size: 0.8125rem;
    font-weight: 500;
    letter-spacing: normal;
    transition: background-color 120ms ease, border-color 120ms ease, color 120ms ease;
}

.building-type-pill:hover {
    background-color: rgba(var(--v-theme-on-surface), 0.05);
    color: rgba(var(--v-theme-on-surface), 0.92);
}

.building-type-pill:focus-visible {
    outline: 2px solid rgb(var(--v-theme-primary));
    outline-offset: 2px;
}

.building-type-pill--active {
    border-color: transparent;
    background-color: rgb(var(--v-theme-secondary));
    color: rgb(var(--v-theme-on-secondary));
    font-weight: 700;
}

.building-type-pill--active:hover {
    background-color: color-mix(in srgb, rgb(var(--v-theme-secondary)) 88%, #fff);
    color: rgb(var(--v-theme-on-secondary));
}

.building-type-pill__label {
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.building-type-pill__icon {
    flex-shrink: 0;
    opacity: 0.9;
}

.building-type-menu {
    min-width: 180px;
    padding: 8px;
    border-radius: 14px;
}

:deep(.building-type-menu-surface) {
    border-radius: 16px;
    overflow: hidden;
}

:deep(.building-type-menu .v-list-item) {
    border-radius: 8px;
}

</style>
