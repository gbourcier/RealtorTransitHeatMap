<script setup lang="ts">
import type { SortBy, SortDir } from "../api/listings";
import type { SortOption } from "../composables/useListings";

interface Props {
    total: number;
    sortBy: SortBy;
    sortDir: SortDir;
    sortOptions: SortOption[];
    mobile?: boolean;
}

defineProps<Props>();

defineEmits<{
    selectSort: [opt: SortOption];
}>();
</script>

<template>
    <div class="list-toolbar" :class="{ 'list-toolbar--mobile': mobile }">
        <div class="list-toolbar__row">
            <div class="list-toolbar__count">
                <span class="list-toolbar__count-num">{{ total.toLocaleString() }}</span>
                <span class="list-toolbar__count-label">listing{{ total === 1 ? "" : "s" }}</span>
            </div>
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
        </div>
    </div>
</template>

<style scoped>
.list-toolbar {
    display: flex;
    flex-direction: column;
    padding: 10px 14px;
    border-bottom: 1px solid rgba(var(--v-theme-on-surface), 0.08);
    flex: 0 0 auto;
}

.list-toolbar--mobile {
    padding: 10px 14px;
    border-bottom: 1px solid rgba(var(--v-theme-on-surface), 0.06);
}

.list-toolbar__row {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 12px;
    min-width: 0;
}

.list-toolbar__count {
    display: flex;
    flex-direction: column;
    line-height: 1;
}

.list-toolbar__count-num {
    font-size: 1rem;
    font-weight: 700;
    letter-spacing: -0.01em;
}

.list-toolbar__count-label {
    font-size: 0.6875rem;
    color: rgba(var(--v-theme-on-surface), 0.55);
    margin-top: 2px;
}

.sort-tabs {
    display: flex;
    align-items: stretch;
    flex: 0 1 auto;
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
    outline: 2px solid rgb(var(--v-theme-secondary));
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
</style>
