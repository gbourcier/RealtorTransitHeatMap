<script setup lang="ts">
import { ref, computed } from "vue";
import type { ListingFiltersState } from "../composables/useListingFilters";
import type { SavedViewsState } from "../composables/useSavedViews";
import type { SavedFilter } from "../api/savedFilters";

interface Props {
    views: SavedViewsState;
    filters: ListingFiltersState;
}

const props = defineProps<Props>();
const emit = defineEmits<{ save: [] }>();

const menuOpen = ref(false);

const favActive = computed(() => props.filters.favoritesOnly.value);
const activeView = computed(() => props.views.active.value);

const triggerState = computed(() =>
    favActive.value ? "fav" : activeView.value ? "saved" : "all",
);
const triggerLabel = computed(() =>
    favActive.value
        ? "Favourites"
        : activeView.value
          ? activeView.value.name
          : "All listings",
);
const triggerIcon = computed(() =>
    favActive.value
        ? "mdi-heart"
        : activeView.value
          ? props.views.isDefault(activeView.value.id)
              ? "mdi-star"
              : "mdi-bookmark"
          : "mdi-format-list-bulleted",
);

function onSelectAll(): void {
    props.views.selectAll();
    menuOpen.value = false;
}

function onSelectFavourites(): void {
    props.views.selectFavourites();
    menuOpen.value = false;
}

function onApply(preset: SavedFilter): void {
    props.views.applySaved(preset);
    menuOpen.value = false;
}

function onSave(): void {
    menuOpen.value = false;
    emit("save");
}

async function onStar(id: string): Promise<void> {
    try {
        await props.views.toggleDefault(id);
    } catch {
        /* owner-scoped toggle; ignore transient failure */
    }
}

async function onDelete(id: string): Promise<void> {
    try {
        await props.views.remove(id);
    } catch {
        /* owner-scoped delete; ignore transient failure */
    }
}
</script>

<template>
    <v-menu
        v-model="menuOpen"
        location="bottom end"
        offset="10"
        :close-on-content-click="false"
        transition="scale-transition"
    >
        <template #activator="{ props: activatorProps }">
            <button
                v-bind="activatorProps"
                type="button"
                class="viewsel"
                :class="`viewsel--${triggerState}`"
                aria-haspopup="listbox"
                :aria-expanded="menuOpen"
                aria-label="Select view"
            >
                <span class="viewsel__content">
                    <v-icon :icon="triggerIcon" size="14" class="viewsel__ic" />
                    <span class="viewsel__label">{{ triggerLabel }}</span>
                    <v-icon icon="mdi-chevron-down" size="12" class="viewsel__caret" />
                </span>
            </button>
        </template>

        <div class="view-menu" role="listbox" aria-label="Views">
            <div class="view-menu__label">Show</div>

            <button
                type="button"
                class="view-menu__item"
                :class="{ 'view-menu__item--on': triggerState === 'all' }"
                @click="onSelectAll"
            >
                <span class="view-menu__ic">
                    <v-icon size="14">mdi-format-list-bulleted</v-icon>
                </span>
                <span class="view-menu__name">All listings</span>
                <v-icon v-if="triggerState === 'all'" size="14" class="view-menu__check">mdi-check</v-icon>
            </button>

            <button
                type="button"
                class="view-menu__item view-menu__item--fav"
                :class="{ 'view-menu__item--on': triggerState === 'fav' }"
                @click="onSelectFavourites"
            >
                <span class="view-menu__ic"><v-icon size="14">mdi-heart</v-icon></span>
                <span class="view-menu__name">Favourites</span>
                <v-icon v-if="triggerState === 'fav'" size="14" class="view-menu__check">mdi-check</v-icon>
            </button>

            <div class="view-menu__div" />
            <div class="view-menu__label">Saved filters</div>

            <div
                v-for="preset in views.sortedList.value"
                :key="preset.id"
                class="view-menu__item view-menu__item--saved"
                :class="{ 'view-menu__item--on': preset.id === views.activeId.value }"
                role="option"
                :aria-selected="preset.id === views.activeId.value"
                tabindex="0"
                @click="onApply(preset)"
                @keydown.enter="onApply(preset)"
            >
                <span
                    class="view-menu__star"
                    :class="{ 'view-menu__star--on': views.isDefault(preset.id) }"
                    role="button"
                    :aria-label="views.isDefault(preset.id) ? 'Unset as default' : 'Set as default'"
                    @click.stop="onStar(preset.id)"
                >
                    <v-icon size="14">
                        {{ views.isDefault(preset.id) ? "mdi-star" : "mdi-star-outline" }}
                    </v-icon>
                </span>
                <span class="view-menu__name">{{ preset.name }}</span>
                <v-icon
                    v-if="preset.id === views.activeId.value"
                    size="14"
                    class="view-menu__check"
                >mdi-check</v-icon>
                <button
                    type="button"
                    class="view-menu__del"
                    aria-label="Delete saved filter"
                    @click.stop="onDelete(preset.id)"
                >
                    <v-icon size="14">mdi-trash-can-outline</v-icon>
                </button>
            </div>

            <div v-if="!views.sortedList.value.length" class="view-menu__empty">
                No saved filters yet — Save below.
            </div>

            <div class="view-menu__div" />
            <button type="button" class="view-menu__item view-menu__save" @click="onSave">
                <span class="view-menu__ic"><v-icon size="15">mdi-plus</v-icon></span>
                <span class="view-menu__name">Save current filters…</span>
            </button>
        </div>
    </v-menu>
</template>

<style scoped>
.viewsel {
    appearance: none;
    -webkit-appearance: none;
    box-sizing: border-box;
    position: relative;
    display: inline-flex;
    align-items: center;
    justify-content: center;
    flex: 0 0 auto;
    height: 28px;
    max-width: 200px;
    padding: 0 10px 0 11px;
    border-radius: 999px;
    white-space: nowrap;
    background: transparent;
    border: 1.5px solid rgba(var(--v-theme-on-surface), 0.24);
    color: rgba(var(--v-theme-on-surface), 0.82);
    font-family: inherit;
    font-size: 0.8125rem;
    font-weight: 600;
    line-height: 1;
    letter-spacing: normal;
    text-align: center;
    cursor: pointer;
    transition: background-color 120ms ease, border-color 120ms ease,
        color 120ms ease;
}

.viewsel:hover {
    border-color: rgba(var(--v-theme-on-surface), 0.42);
    background: rgba(var(--v-theme-on-surface), 0.04);
}

.viewsel:focus-visible {
    outline: none;
}

.viewsel:focus-visible::before {
    content: "";
    position: absolute;
    inset: -4px;
    border: 2px solid rgb(var(--v-theme-primary));
    border-radius: inherit;
    pointer-events: none;
}

.viewsel__content {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 7px;
    min-width: 0;
}

.viewsel__ic {
    flex: 0 0 auto;
    opacity: 0.9;
}

.viewsel__label {
    overflow: hidden;
    text-overflow: ellipsis;
}

.viewsel__caret {
    flex: 0 0 auto;
    opacity: 0.65;
}

.viewsel--saved,
.viewsel--saved:hover {
    border-width: 2px;
    border-color: rgb(var(--v-theme-primary));
    color: rgb(var(--v-theme-primary));
    background: rgba(var(--v-theme-primary), 0.13);
    padding: 0 9px 0 10px;
}

.viewsel--fav,
.viewsel--fav:hover {
    border-width: 2px;
    border-color: rgb(var(--v-theme-accent));
    color: rgb(var(--v-theme-accent));
    background: rgba(var(--v-theme-accent), 0.15);
    padding: 0 9px 0 10px;
}

.viewsel--saved::after,
.viewsel--fav::after {
    content: "";
    position: absolute;
    inset: -4px;
    border-radius: inherit;
    pointer-events: none;
}

.viewsel--saved::after {
    border: 3px solid rgba(var(--v-theme-primary), 0.1);
}

.viewsel--fav::after {
    border: 3px solid rgba(var(--v-theme-accent), 0.1);
}

.viewsel--saved .viewsel__ic,
.viewsel--saved .viewsel__caret,
.viewsel--fav .viewsel__ic,
.viewsel--fav .viewsel__caret {
    opacity: 1;
}

.view-menu {
    width: 290px;
    padding: 6px;
    border-radius: 14px;
    background: rgb(var(--v-theme-surface));
    border: 1px solid rgba(var(--v-theme-on-surface), 0.1);
    box-shadow: 0 20px 56px rgba(var(--v-theme-shadow), 0.55);
    color: rgba(var(--v-theme-on-surface), 0.92);
}

.view-menu__label {
    font-size: 0.65625rem;
    font-weight: 700;
    letter-spacing: 1.2px;
    text-transform: uppercase;
    color: rgba(var(--v-theme-on-surface), 0.34);
    padding: 9px 11px 6px;
}

.view-menu__item {
    display: flex;
    align-items: center;
    gap: 10px;
    width: 100%;
    height: 40px;
    padding: 0 11px;
    border: 0;
    border-radius: 9px;
    background: transparent;
    color: rgba(var(--v-theme-on-surface), 0.86);
    font-size: 0.84375rem;
    font-weight: 500;
    text-align: left;
    cursor: pointer;
    transition: background-color 120ms ease;
}

.view-menu__item:hover {
    background: rgba(var(--v-theme-on-surface), 0.06);
}

.view-menu__item:focus-visible {
    outline: 2px solid rgb(var(--v-theme-primary));
    outline-offset: -2px;
}

.view-menu__ic {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 16px;
    flex: 0 0 auto;
    color: rgba(var(--v-theme-on-surface), 0.5);
}

.view-menu__star {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 16px;
    flex: 0 0 auto;
    color: rgba(var(--v-theme-on-surface), 0.4);
    border-radius: 4px;
    cursor: pointer;
    transition: color 120ms ease, transform 80ms ease;
}

.view-menu__star:hover {
    color: rgb(var(--v-theme-primary));
    transform: scale(1.18);
}

.view-menu__star--on {
    color: rgb(var(--v-theme-primary));
}

.view-menu__name {
    flex: 1 1 auto;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.view-menu__check {
    flex: 0 0 auto;
    color: rgb(var(--v-theme-primary));
}

.view-menu__item--on {
    background: rgba(var(--v-theme-primary), 0.1);
}

.view-menu__item--on .view-menu__name {
    color: rgb(var(--v-theme-primary));
    font-weight: 700;
}

.view-menu__item--fav .view-menu__ic {
    color: rgb(var(--v-theme-accent));
}

.view-menu__item--fav.view-menu__item--on {
    background: rgba(var(--v-theme-accent), 0.12);
}

.view-menu__item--fav.view-menu__item--on .view-menu__name {
    color: rgb(var(--v-theme-accent));
}

.view-menu__item--fav.view-menu__item--on .view-menu__check {
    color: rgb(var(--v-theme-accent));
}

.view-menu__del {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 26px;
    height: 26px;
    flex: 0 0 auto;
    border: 0;
    border-radius: 7px;
    background: transparent;
    color: rgba(var(--v-theme-on-surface), 0.4);
    opacity: 0;
    cursor: pointer;
    transition: opacity 120ms ease, background-color 120ms ease, color 120ms ease;
}

.view-menu__item--saved:hover .view-menu__del,
.view-menu__item--saved:focus-within .view-menu__del {
    opacity: 1;
}

.view-menu__del:hover {
    background: rgba(var(--v-theme-error), 0.16);
    color: rgb(var(--v-theme-error));
}

.view-menu__div {
    height: 1px;
    background: rgba(var(--v-theme-on-surface), 0.08);
    margin: 6px 8px;
}

.view-menu__empty {
    padding: 6px 12px 10px;
    font-size: 0.78125rem;
    color: rgba(var(--v-theme-on-surface), 0.42);
}

.view-menu__save .view-menu__ic {
    color: rgb(var(--v-theme-primary));
}

.view-menu__save .view-menu__name {
    font-weight: 600;
}

@media (max-width: 600px) {
    .viewsel {
        width: 58px;
        max-width: none;
        padding: 0;
    }

    .viewsel--saved,
    .viewsel--fav {
        padding: 0;
    }

    .viewsel__label {
        display: none;
    }

    .viewsel__content {
        position: absolute;
        left: 50%;
        top: 50%;
        display: grid;
        grid-template-columns: 14px 12px;
        column-gap: 8px;
        flex: none;
        width: 34px;
        transform: translate(-50%, -50%);
    }
}
</style>
