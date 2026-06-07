<script setup lang="ts">
import { ref, computed } from "vue";
import { useDisplay } from "vuetify";
import type { ListingFiltersState } from "../composables/useListingFilters";
import type { SavedViewsState } from "../composables/useSavedViews";
import type { SavedFilter } from "../api/savedFilters";

interface Props {
    views: SavedViewsState;
    filters: ListingFiltersState;
}

const props = defineProps<Props>();
const emit = defineEmits<{ save: [] }>();

const { mobile } = useDisplay();
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
          : "Saved views",
);
const triggerIcon = computed(() =>
    favActive.value
        ? "mdi-heart"
        : activeView.value
          ? props.views.isDefault(activeView.value.id)
              ? "mdi-star"
              : "mdi-star-outline"
          : "mdi-star-outline",
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
        :location="mobile ? 'bottom center' : 'bottom end'"
        offset="10"
        :close-on-content-click="false"
        :transition="mobile ? 'mobile-sheet-transition' : 'scale-transition'"
        :content-class="mobile ? 'mobile-sheet-menu mobile-sheet-menu--views' : undefined"
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
                    <v-icon :icon="triggerIcon" size="16" class="viewsel__ic" />
                    <span class="viewsel__label">{{ triggerLabel }}</span>
                    <v-icon icon="mdi-chevron-down" size="15" class="viewsel__caret" />
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
                    <v-icon size="18">mdi-format-list-bulleted</v-icon>
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
                <span class="view-menu__ic"><v-icon size="18">mdi-heart</v-icon></span>
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
    height: 40px;
    max-width: 230px;
    padding: 0 14px;
    border-radius: 999px;
    white-space: nowrap;
    background: #2a2d27;
    border: 1px solid rgba(244, 241, 232, 0.12);
    color: #f4f1e8;
    font-family: inherit;
    font-size: 14px;
    font-weight: 600;
    line-height: 1;
    letter-spacing: 0;
    text-align: center;
    cursor: pointer;
    transition: background-color 140ms ease, border-color 140ms ease,
        box-shadow 140ms ease, transform 60ms ease;
}

.viewsel:hover {
    border-color: rgba(244, 241, 232, 0.22);
    background: #34382f;
}

.viewsel:active {
    transform: translateY(1px);
}

.viewsel:focus-visible {
    outline: none;
}

.viewsel:focus-visible::before {
    content: "";
    position: absolute;
    inset: -4px;
    border: 2px solid #6ccff6;
    border-radius: inherit;
    pointer-events: none;
}

.viewsel__content {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    min-width: 0;
}

.viewsel__ic {
    flex: 0 0 auto;
    color: #f2c14e;
}

.viewsel__label {
    overflow: hidden;
    text-overflow: ellipsis;
}

.viewsel__caret {
    flex: 0 0 auto;
    color: rgba(244, 241, 232, 0.52);
}

.viewsel--saved,
.viewsel--saved:hover {
    border: 1.5px solid #6ccff6;
    color: #f4f1e8;
    background: #2a2d27;
    box-shadow: 0 0 0 3px rgba(108, 207, 246, 0.15);
}

.viewsel--fav,
.viewsel--fav:hover {
    border: 1.5px solid #6ccff6;
    color: #f4f1e8;
    background: #2a2d27;
    box-shadow: 0 0 0 3px rgba(108, 207, 246, 0.15);
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
    border: 0;
}

.viewsel--fav::after {
    border: 0;
}

.viewsel--saved .viewsel__ic,
.viewsel--saved .viewsel__caret,
.viewsel--fav .viewsel__ic,
.viewsel--fav .viewsel__caret {
    opacity: 1;
}

.view-menu {
    width: 300px;
    padding: 8px;
    border-radius: 16px;
    background: #262925;
    border: 1px solid rgba(244, 241, 232, 0.12);
    box-shadow: 0 24px 60px -18px rgba(0, 0, 0, 0.8);
    color: #f4f1e8;
    font-family: Inter, system-ui, sans-serif;
}

.view-menu__label {
    font-size: 11px;
    font-weight: 600;
    letter-spacing: 0.13em;
    text-transform: uppercase;
    color: rgba(244, 241, 232, 0.34);
    padding: 10px 12px 6px;
}

.view-menu__item {
    display: flex;
    align-items: center;
    gap: 13px;
    width: 100%;
    height: 46px;
    padding: 0 12px;
    border: 0;
    border-radius: 11px;
    background: transparent;
    color: #f4f1e8;
    font-size: 15px;
    font-weight: 600;
    text-align: left;
    cursor: pointer;
    transition: background-color 120ms ease;
}

.view-menu__item:hover {
    background: rgba(244, 241, 232, 0.06);
}

.view-menu__item:focus-visible {
    outline: 2px solid #6ccff6;
    outline-offset: -2px;
}

.view-menu__ic {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 16px;
    flex: 0 0 auto;
    color: rgba(244, 241, 232, 0.52);
}

.view-menu__star {
    display: inline-flex;
    align-items: center;
    justify-content: center;
    width: 16px;
    flex: 0 0 auto;
    color: rgba(244, 241, 232, 0.34);
    border-radius: 4px;
    cursor: pointer;
    transition: color 120ms ease, transform 80ms ease;
}

.view-menu__star:hover {
    color: #b6f24a;
    transform: scale(1.18);
}

.view-menu__star--on {
    color: #f2c14e;
}

.view-menu__name {
    flex: 1 1 auto;
    overflow: hidden;
    text-overflow: ellipsis;
    white-space: nowrap;
}

.view-menu__check {
    flex: 0 0 auto;
    color: #b6f24a;
}

.view-menu__item--on {
    background: rgba(182, 242, 74, 0.12);
    color: #b6f24a;
}

.view-menu__item--on .view-menu__name {
    color: #b6f24a;
    font-weight: 700;
}

.view-menu__item--fav .view-menu__ic {
    color: #f2c14e;
}

.view-menu__item--fav.view-menu__item--on {
    background: rgba(182, 242, 74, 0.12);
}

.view-menu__item--fav.view-menu__item--on .view-menu__name {
    color: #b6f24a;
}

.view-menu__item--fav.view-menu__item--on .view-menu__check {
    color: #b6f24a;
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
    color: rgba(244, 241, 232, 0.34);
    opacity: 0;
    cursor: pointer;
    transition: opacity 120ms ease, background-color 120ms ease, color 120ms ease;
}

.view-menu__item--saved:hover .view-menu__del,
.view-menu__item--saved:focus-within .view-menu__del {
    opacity: 1;
}

.view-menu__del:hover {
    background: rgba(255, 92, 138, 0.16);
    color: #ff5c8a;
}

.view-menu__div {
    height: 1px;
    background: rgba(244, 241, 232, 0.12);
    margin: 7px 6px;
}

.view-menu__empty {
    padding: 6px 12px 10px;
    font-size: 0.78125rem;
    color: rgba(244, 241, 232, 0.52);
}

.view-menu__save .view-menu__ic {
    color: #b6f24a;
}

.view-menu__save .view-menu__name {
    font-weight: 600;
}

@media (max-width: 899px) {
    :global(.mobile-sheet-menu.v-overlay__content) {
        position: fixed !important;
        top: auto !important;
        right: 11px !important;
        bottom: 0 !important;
        left: 11px !important;
        width: auto !important;
        min-width: 0 !important;
        max-width: none !important;
    }

    :global(.mobile-sheet-transition-enter-active),
    :global(.mobile-sheet-transition-leave-active) {
        transition: opacity 180ms ease, transform 260ms cubic-bezier(0.22, 0.7, 0.3, 1);
    }

    :global(.mobile-sheet-transition-enter-from),
    :global(.mobile-sheet-transition-leave-to) {
        opacity: 0;
        transform: translateY(100%);
    }

    .viewsel {
        width: auto;
        min-width: 58px;
        height: 38px;
        max-width: none;
        padding: 0 11px;
    }

    .viewsel--saved,
    .viewsel--fav {
        padding: 0 11px;
    }

    .viewsel__label {
        display: none;
    }

    .viewsel__content {
        position: absolute;
        left: 50%;
        top: 50%;
        display: grid;
        grid-template-columns: 16px 14px;
        column-gap: 8px;
        width: 38px;
        transform: translate(-50%, -50%);
    }

    .view-menu {
        width: 100%;
        max-height: 88dvh;
        overflow-y: auto;
        border-radius: 22px 22px 0 0;
    }

    .view-menu__del {
        opacity: 1;
    }
}
</style>
