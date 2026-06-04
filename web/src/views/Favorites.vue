<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import { useDisplay } from "vuetify";
import {
    listFavorites,
    removeFavoritesBatch,
    type Favorite,
    type FavoriteSortBy,
} from "../api/favorites";
import { useInfiniteScroll } from "../composables/useInfiniteScroll";
import {
    formatPrice,
    formatDate,
    formatCommute,
    parseAddress,
} from "../utils/listingFormat";
import EmptyState from "../components/EmptyState.vue";

defineOptions({ name: "Favorites" });

const { mobile } = useDisplay();

const items = ref<Favorite[]>([]);
const total = ref(0);
const loading = ref(false);
const sortBy = ref<FavoriteSortBy>("favorited_date");
const sortDir = ref<"asc" | "desc">("desc");
const selected = ref<Set<string>>(new Set());

const errorToast = ref<string | null>(null);
const errorToastOpen = ref(false);

const confirmOpen = ref(false);
const removing = ref(false);

const bodyEl = ref<HTMLElement | null>(null);
const sentinelEl = ref<HTMLElement | null>(null);

const limit = 50;
let loadGen = 0;
let inFlight: Promise<void> | null = null;

const hasMore = computed(() => items.value.length < total.value);
const selectedCount = computed(() => selected.value.size);
const allLoadedSelected = computed(
    () =>
        items.value.length > 0 &&
        items.value.every((f) => selected.value.has(rowKey(f))),
);

const columns: { key: FavoriteSortBy; label: string; numeric: boolean }[] = [
    { key: "price", label: "Price", numeric: true },
    { key: "commute", label: "Commute", numeric: true },
    { key: "favorited_date", label: "Favorited", numeric: false },
    { key: "listing_posted_date", label: "Listed", numeric: false },
];

const activeSortLabel = computed(
    () => columns.find((c) => c.key === sortBy.value)?.label ?? "",
);

function rowKey(f: Favorite): string {
    return `${f.board}-${f.mls}`;
}

function showError(message: string): void {
    errorToast.value = message;
    errorToastOpen.value = true;
}

function loadMore(gen: number = loadGen): Promise<void> {
    if (gen !== loadGen) return Promise.resolve();
    if (inFlight) return inFlight.then(() => loadMore(gen));
    if (items.value.length > 0 && items.value.length >= total.value) {
        return Promise.resolve();
    }
    loading.value = true;
    inFlight = (async () => {
        try {
            const res = await listFavorites({
                limit,
                offset: items.value.length,
                sortBy: sortBy.value,
                sortDir: sortDir.value,
            });
            if (gen !== loadGen) return;
            items.value = [...items.value, ...res.items];
            total.value = res.total;
        } catch (e: any) {
            if (gen === loadGen) {
                showError(
                    e?.response?.data?.error ??
                        e?.message ??
                        "Failed to load favorites.",
                );
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
    selected.value = new Set();
    await loadMore(gen);
}

function setSort(col: FavoriteSortBy): void {
    if (sortBy.value === col) {
        sortDir.value = sortDir.value === "asc" ? "desc" : "asc";
    } else {
        sortBy.value = col;
        sortDir.value = col === "price" || col === "commute" ? "asc" : "desc";
    }
    loadInitial();
}

function toggleRow(f: Favorite): void {
    const k = rowKey(f);
    const next = new Set(selected.value);
    if (next.has(k)) next.delete(k);
    else next.add(k);
    selected.value = next;
}

function toggleAll(): void {
    selected.value = allLoadedSelected.value
        ? new Set()
        : new Set(items.value.map(rowKey));
}

async function confirmRemove(): Promise<void> {
    removing.value = true;
    try {
        const toRemove = items.value
            .filter((f) => selected.value.has(rowKey(f)))
            .map((f) => ({ board: f.board, mls: f.mls }));
        await removeFavoritesBatch(toRemove);
        confirmOpen.value = false;
        await loadInitial();
    } catch (e: any) {
        showError(e?.response?.data?.error ?? "Failed to remove favorites.");
    } finally {
        removing.value = false;
    }
}

function openListing(f: Favorite): void {
    if (f.slug) window.open(f.slug, "_blank", "noopener,noreferrer");
}

function mapsUrl(f: Favorite): string {
    return `https://www.google.com/maps/search/?api=1&query=${f.latitude},${f.longitude}`;
}

useInfiniteScroll(sentinelEl, bodyEl, () => loadMore());

onMounted(loadInitial);
</script>

<template>
    <v-container fluid class="pa-2 pa-sm-6">
        <v-card class="favorites-card">
            <v-card-title class="d-flex align-center">
                <span>Manage Favorites</span>
                <span v-if="total > 0" class="favorites-count">{{ total }}</span>
                <v-spacer />
                <v-btn
                    color="error"
                    variant="tonal"
                    size="small"
                    prepend-icon="mdi-delete-outline"
                    :disabled="selectedCount === 0"
                    @click="confirmOpen = true"
                >
                    Remove{{ selectedCount > 0 ? ` (${selectedCount})` : "" }}
                </v-btn>
            </v-card-title>

            <v-divider />

            <div
                v-if="mobile && (items.length > 0 || loading)"
                class="favorites-mobile-bar"
            >
                <v-checkbox-btn
                    :model-value="allLoadedSelected"
                    density="compact"
                    hide-details
                    label="Select all"
                    class="favorites-mobile-bar__check"
                    @update:model-value="toggleAll"
                />
                <v-spacer />
                <v-menu>
                    <template #activator="{ props }">
                        <v-btn
                            v-bind="props"
                            variant="tonal"
                            size="small"
                            :prepend-icon="sortDir === 'asc' ? 'mdi-sort-ascending' : 'mdi-sort-descending'"
                            append-icon="mdi-menu-down"
                        >
                            {{ activeSortLabel }}
                        </v-btn>
                    </template>
                    <v-list density="compact">
                        <v-list-item
                            v-for="col in columns"
                            :key="col.key"
                            :active="sortBy === col.key"
                            @click="setSort(col.key)"
                        >
                            <v-list-item-title>{{ col.label }}</v-list-item-title>
                            <template #append>
                                <v-icon v-if="sortBy === col.key" size="16">
                                    {{ sortDir === "asc" ? "mdi-arrow-up" : "mdi-arrow-down" }}
                                </v-icon>
                            </template>
                        </v-list-item>
                    </v-list>
                </v-menu>
            </div>

            <div ref="bodyEl" class="favorites-body">
                <div v-if="mobile && (items.length > 0 || loading)" class="favorites-cards">
                    <div
                        v-for="f in items"
                        :key="rowKey(f)"
                        class="favorites-card-item"
                        :class="{ 'favorites-card-item--selected': selected.has(rowKey(f)) }"
                    >
                        <v-checkbox-btn
                            :model-value="selected.has(rowKey(f))"
                            density="compact"
                            hide-details
                            class="favorites-card-item__check"
                            @update:model-value="toggleRow(f)"
                        />
                        <div class="favorites-card-item__info">
                            <div class="favorites-card-item__top">
                                <div class="favorites-card-item__title">
                                    <div class="favorites-listing__street">
                                        {{ parseAddress(f.address).street }}
                                        <v-chip
                                            v-if="!f.isAvailable"
                                            size="x-small"
                                            color="warning"
                                            variant="tonal"
                                            class="ms-1"
                                        >expired</v-chip>
                                    </div>
                                    <div
                                        v-if="parseAddress(f.address).locality"
                                        class="favorites-listing__locality"
                                    >
                                        {{ parseAddress(f.address).locality }}
                                    </div>
                                </div>
                                <div class="favorites-card-item__figures">
                                    <div class="favorites-table__price">
                                        {{ formatPrice(f.currentPrice) }}
                                    </div>
                                    <div class="favorites-card-item__commute">
                                        {{ formatCommute(f.commuteSecondsDowntown) }}
                                    </div>
                                </div>
                            </div>
                            <div class="favorites-card-item__foot">
                                <span class="favorites-card-item__dates">
                                    Favorited {{ formatDate(f.favoritedAt) }} · Listed
                                    {{ formatDate(f.firstSeenAt) }}
                                </span>
                                <div class="favorites-card-item__actions">
                                    <v-btn
                                        icon="mdi-open-in-new"
                                        variant="text"
                                        size="small"
                                        aria-label="View listing"
                                        @click="openListing(f)"
                                    />
                                    <v-btn
                                        icon="mdi-map-marker-outline"
                                        variant="text"
                                        size="small"
                                        aria-label="View on Google Maps"
                                        :href="mapsUrl(f)"
                                        target="_blank"
                                        rel="noopener noreferrer"
                                    />
                                </div>
                            </div>
                        </div>
                    </div>
                </div>

                <table v-else-if="items.length > 0 || loading" class="favorites-table">
                    <thead>
                        <tr>
                            <th class="favorites-table__check">
                                <v-checkbox-btn
                                    :model-value="allLoadedSelected"
                                    density="compact"
                                    hide-details
                                    @update:model-value="toggleAll"
                                />
                            </th>
                            <th class="favorites-table__listing">Listing</th>
                            <th
                                v-for="col in columns"
                                :key="col.key"
                                class="favorites-table__sortable"
                                :class="{ 'favorites-table__num': col.numeric }"
                            >
                                <button
                                    type="button"
                                    class="favorites-sort"
                                    :class="{ 'favorites-sort--active': sortBy === col.key }"
                                    @click="setSort(col.key)"
                                >
                                    <span>{{ col.label }}</span>
                                    <v-icon
                                        size="16"
                                        class="favorites-sort__icon"
                                        :class="{ 'favorites-sort__icon--hidden': sortBy !== col.key }"
                                    >
                                        {{ sortBy === col.key && sortDir === "asc" ? "mdi-arrow-up" : "mdi-arrow-down" }}
                                    </v-icon>
                                </button>
                            </th>
                            <th class="favorites-table__actions">Actions</th>
                        </tr>
                    </thead>
                    <tbody>
                        <tr
                            v-for="f in items"
                            :key="rowKey(f)"
                            class="favorites-row"
                            :class="{ 'favorites-row--selected': selected.has(rowKey(f)) }"
                        >
                            <td class="favorites-table__check">
                                <v-checkbox-btn
                                    :model-value="selected.has(rowKey(f))"
                                    density="compact"
                                    hide-details
                                    @update:model-value="toggleRow(f)"
                                />
                            </td>
                            <td class="favorites-table__listing">
                                <div class="favorites-listing__street">
                                    {{ parseAddress(f.address).street }}
                                    <v-chip
                                        v-if="!f.isAvailable"
                                        size="x-small"
                                        color="warning"
                                        variant="tonal"
                                        class="ms-1"
                                    >expired</v-chip>
                                </div>
                                <div
                                    v-if="parseAddress(f.address).locality"
                                    class="favorites-listing__locality"
                                >
                                    {{ parseAddress(f.address).locality }}
                                </div>
                            </td>
                            <td class="favorites-table__num favorites-table__price">
                                {{ formatPrice(f.currentPrice) }}
                            </td>
                            <td class="favorites-table__num">
                                {{ formatCommute(f.commuteSecondsDowntown) }}
                            </td>
                            <td>{{ formatDate(f.favoritedAt) }}</td>
                            <td>{{ formatDate(f.firstSeenAt) }}</td>
                            <td class="favorites-table__actions">
                                <v-btn
                                    icon="mdi-open-in-new"
                                    variant="text"
                                    size="small"
                                    aria-label="View listing"
                                    @click="openListing(f)"
                                />
                                <v-btn
                                    icon="mdi-map-marker-outline"
                                    variant="text"
                                    size="small"
                                    aria-label="View on Google Maps"
                                    :href="mapsUrl(f)"
                                    target="_blank"
                                    rel="noopener noreferrer"
                                />
                            </td>
                        </tr>
                    </tbody>
                </table>

                <EmptyState v-else text="No favorites yet." icon="mdi-heart-outline" />

                <div
                    v-if="hasMore"
                    ref="sentinelEl"
                    class="favorites-sentinel"
                >
                    <v-progress-circular v-if="loading" indeterminate size="20" width="2" />
                </div>
            </div>
        </v-card>
    </v-container>

    <v-dialog v-model="confirmOpen" max-width="420">
        <v-card>
            <v-card-title>Remove favorites</v-card-title>
            <v-card-text>
                Remove {{ selectedCount }} favorite{{ selectedCount === 1 ? "" : "s" }}? This
                cannot be undone.
            </v-card-text>
            <v-card-actions class="pb-4 px-6">
                <v-spacer />
                <v-btn variant="text" @click="confirmOpen = false">Cancel</v-btn>
                <v-btn color="error" :loading="removing" @click="confirmRemove">Remove</v-btn>
            </v-card-actions>
        </v-card>
    </v-dialog>

    <v-snackbar v-model="errorToastOpen" color="error" location="top" :timeout="6000">
        {{ errorToast }}
        <template #actions>
            <v-btn variant="text" @click="errorToastOpen = false">Dismiss</v-btn>
        </template>
    </v-snackbar>
</template>

<style scoped>
.favorites-card {
    max-width: 1100px;
    margin: 0 auto;
}

.favorites-count {
    margin-left: 10px;
    font-size: 0.8125rem;
    font-weight: 500;
    color: rgba(var(--v-theme-on-surface), 0.55);
}

.favorites-body {
    max-height: calc(100dvh - 200px);
    overflow-y: auto;
}

.favorites-table {
    width: 100%;
    border-collapse: collapse;
    font-size: 0.875rem;
}

.favorites-table thead th {
    position: sticky;
    top: 0;
    z-index: 1;
    background-color: rgb(var(--v-theme-surface));
    text-align: left;
    font-weight: 600;
    color: rgba(var(--v-theme-on-surface), 0.7);
    padding: 10px 12px;
    border-bottom: 1px solid rgba(var(--v-theme-on-surface), 0.1);
    white-space: nowrap;
}

.favorites-table__check {
    width: 44px;
    padding-left: 12px;
}

.favorites-table__actions {
    width: 96px;
    white-space: nowrap;
}

.favorites-sort {
    display: inline-flex;
    align-items: center;
    gap: 4px;
    background: transparent;
    border: 0;
    padding: 0;
    font: inherit;
    font-weight: 600;
    color: rgba(var(--v-theme-on-surface), 0.7);
    cursor: pointer;
}

.favorites-sort__icon--hidden {
    visibility: hidden;
}

.favorites-sort:hover {
    color: rgba(var(--v-theme-on-surface), 0.95);
}

.favorites-sort--active {
    color: rgb(var(--v-theme-primary));
}

.favorites-row td {
    padding: 10px 12px;
    border-bottom: 1px solid rgba(var(--v-theme-on-surface), 0.06);
    vertical-align: middle;
}

.favorites-row--selected td {
    background-color: rgba(var(--v-theme-primary), 0.08);
}

.favorites-listing__street {
    font-weight: 500;
}

.favorites-listing__locality {
    font-size: 0.75rem;
    color: rgba(var(--v-theme-on-surface), 0.55);
    margin-top: 2px;
}

.favorites-table__price {
    font-weight: 600;
}

.favorites-sentinel {
    display: flex;
    align-items: center;
    justify-content: center;
    min-height: 36px;
    padding: 8px 0;
}

.favorites-mobile-bar {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 8px 12px;
    border-bottom: 1px solid rgba(var(--v-theme-on-surface), 0.1);
}

.favorites-mobile-bar__check {
    flex: 0 0 auto;
}

.favorites-mobile-bar__check :deep(.v-label) {
    white-space: nowrap;
    opacity: 0.85;
}

.favorites-cards {
    display: flex;
    flex-direction: column;
}

.favorites-card-item {
    display: flex;
    align-items: flex-start;
    gap: 8px;
    padding: 12px 12px 6px;
    border-bottom: 1px solid rgba(var(--v-theme-on-surface), 0.06);
}

.favorites-card-item--selected {
    background-color: rgba(var(--v-theme-primary), 0.08);
}

.favorites-card-item__check {
    flex: 0 0 auto;
    margin-top: 1px;
}

.favorites-card-item__info {
    flex: 1 1 auto;
    min-width: 0;
}

.favorites-card-item__top {
    display: flex;
    align-items: flex-start;
    gap: 12px;
}

.favorites-card-item__title {
    flex: 1 1 auto;
    min-width: 0;
}

.favorites-card-item__figures {
    flex: 0 0 auto;
    text-align: right;
}

.favorites-card-item__commute {
    margin-top: 2px;
    font-size: 0.8125rem;
    color: rgba(var(--v-theme-on-surface), 0.7);
}

.favorites-card-item__foot {
    display: flex;
    align-items: center;
    justify-content: space-between;
    gap: 8px;
    margin-top: 4px;
}

.favorites-card-item__dates {
    font-size: 0.75rem;
    color: rgba(var(--v-theme-on-surface), 0.55);
}

.favorites-card-item__actions {
    flex: 0 0 auto;
    display: flex;
    margin-right: -6px;
}
</style>
