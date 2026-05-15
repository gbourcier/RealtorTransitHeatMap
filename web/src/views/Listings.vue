<script setup lang="ts">
import { ref, computed, onMounted } from "vue";
import {
    listListings,
    type Listing,
    type SortBy,
    type SortDir,
} from "../api/listings";

const items = ref<Listing[]>([]);
const total = ref(0);
const limit = ref(50);
const page = ref(1);
const loading = ref(false);
const error = ref<string | null>(null);
const sortBy = ref<SortBy>("first_seen_at");
const sortDir = ref<SortDir>("desc");

const offset = computed(() => (page.value - 1) * limit.value);
const pageCount = computed(() => Math.ceil(total.value / limit.value));

async function load() {
    loading.value = true;
    error.value = null;
    try {
        const res = await listListings({
            limit: limit.value,
            offset: offset.value,
            sortBy: sortBy.value,
            sortDir: sortDir.value,
        });
        items.value = res.items;
        total.value = res.total;
    } catch (e: any) {
        error.value =
            e?.response?.data?.error ?? e?.message ?? "failed to load listings";
    } finally {
        loading.value = false;
    }
}

function toggleSort(col: SortBy) {
    if (sortBy.value === col) {
        sortDir.value = sortDir.value === "asc" ? "desc" : "asc";
    } else {
        sortBy.value = col;
        sortDir.value = "desc";
    }
    page.value = 1;
    load();
}

function sortIcon(col: SortBy): string {
    if (sortBy.value !== col) return "mdi-unfold-more-horizontal";
    return sortDir.value === "asc" ? "mdi-arrow-up" : "mdi-arrow-down";
}

function formatPrice(price: number | null): string {
    if (price == null) return "—";
    return new Intl.NumberFormat("en-CA", {
        style: "currency",
        currency: "CAD",
        maximumFractionDigits: 0,
    }).format(price);
}

function daysSince(unix: number): number {
    const date = new Date(unix * 1000);
    const now = new Date();
    const startOfDate = new Date(date.getFullYear(), date.getMonth(), date.getDate());
    const startOfToday = new Date(now.getFullYear(), now.getMonth(), now.getDate());
    return Math.round((startOfToday.getTime() - startOfDate.getTime()) / 86400000);
}

function formatDate(unix: number): string {
    const diff = daysSince(unix);
    if (diff === 0) return "today";
    if (diff === 1) return "yesterday";
    const date = new Date(unix * 1000);
    const dd = String(date.getDate()).padStart(2, "0");
    const mm = String(date.getMonth() + 1).padStart(2, "0");
    const yyyy = date.getFullYear();
    return `${dd}-${mm}-${yyyy}`;
}

function isNew(unix: number): boolean {
    return daysSince(unix) === 0;
}

function formatCommute(seconds: number | null): string {
    if (seconds == null) return "—";
    return `${Math.round(seconds / 60)} min`;
}

function commuteMapUrl(address: string | null): string | null {
    if (!address) return null;
    const params = new URLSearchParams({
        saddr: address,
        daddr: "McGill Station, Montreal, QC",
        dirflg: "r",
        ttype: "arr",
    });
    return `https://www.google.com/maps?${params.toString()}`;
}

onMounted(load);
</script>

<template>
    <v-container fluid class="pa-6">
        <v-card>
            <v-card-title class="d-flex align-center">
                <span>Listings</span>
                <v-spacer />
                <span v-if="total > 0" class="text-body-2 text-medium-emphasis mr-4">{{ total }} total</span>
                <v-btn variant="text" icon="mdi-refresh" :loading="loading" @click="load" />
            </v-card-title>
            <v-divider />

            <v-alert v-if="error" type="error" variant="tonal" class="ma-3">{{
                error
            }}</v-alert>

            <v-card-text v-if="loading && items.length === 0" class="text-center py-8">
                <v-progress-circular indeterminate />
            </v-card-text>

            <v-table v-else-if="items.length > 0" density="comfortable" class="listings-table">
                <thead>
                    <tr>
                        <th>Address</th>
                        <th class="sortable-col text-right" @click="toggleSort('price')">
                            Price
                            <v-icon size="small" class="sort-icon">{{
                                sortIcon("price")
                            }}</v-icon>
                        </th>
                        <th class="sortable-col" @click="toggleSort('first_seen_at')">
                            First Seen
                            <v-icon size="small" class="sort-icon">{{
                                sortIcon("first_seen_at")
                            }}</v-icon>
                        </th>
                        <th class="sortable-col" @click="toggleSort('commute_time')">
                            Commute Time
                            <v-icon size="small" class="sort-icon">{{
                                sortIcon("commute_time")
                            }}</v-icon>
                        </th>
                    </tr>
                </thead>
                <tbody>
                    <tr v-for="item in items" :key="`${item.board}-${item.mls}`" class="listing-row">
                        <td>
                            <v-tooltip v-if="item.slug && item.address" location="top" open-delay="400">
                                <template #activator="{ props }">
                                    <a v-bind="props" :href="item.slug" target="_blank" rel="noopener noreferrer"
                                        class="address-link">
                                        <span>{{ item.address }}</span>
                                        <v-icon size="small" class="address-link__icon">mdi-open-in-new</v-icon>
                                    </a>
                                </template>
                                <span>View on Realtor.ca</span>
                            </v-tooltip>
                            <template v-else>{{ item.address || "—" }}</template>
                        </td>
                        <td class="text-right">
                            {{ formatPrice(item.currentPrice) }}
                        </td>
                        <td>
                            <span class="first-seen">
                                <span>{{ formatDate(item.firstSeenAt) }}</span>
                                <v-chip v-if="isNew(item.firstSeenAt)" size="x-small" color="secondary"
                                    variant="outlined">new</v-chip>
                            </span>
                        </td>
                        <td>
                            <v-tooltip v-if="
                                item.commuteSecondsDowntown != null &&
                                item.address
                            " location="top" open-delay="400">
                                <template #activator="{ props }">
                                    <a v-bind="props" :href="commuteMapUrl(item.address) ?? '#'" target="_blank"
                                        rel="noopener noreferrer" class="commute-link">
                                        <v-icon size="small" class="commute-link__icon">mdi-directions</v-icon>
                                        <span>{{ formatCommute(item.commuteSecondsDowntown) }}</span>
                                    </a>
                                </template>
                                <span>Get directions to downtown</span>
                            </v-tooltip>
                            <span v-else class="text-medium-emphasis">
                                {{ formatCommute(item.commuteSecondsDowntown) }}
                            </span>
                        </td>
                    </tr>
                </tbody>
            </v-table>

            <v-card-text v-else class="text-medium-emphasis text-center py-8">
                No listings found.
            </v-card-text>

            <template v-if="pageCount > 1">
                <v-divider />
                <v-card-text class="d-flex justify-center pa-3">
                    <v-pagination :model-value="page" :length="pageCount" density="comfortable" @update:model-value="
                        (p) => {
                            page = p;
                            load();
                        }
                    " />
                </v-card-text>
            </template>
        </v-card>
    </v-container>
</template>

<style scoped>
.sortable-col {
    cursor: pointer;
    user-select: none;
    white-space: nowrap;
}

.sortable-col:hover {
    background-color: rgba(var(--v-theme-on-surface), 0.04);
}

.sort-icon {
    opacity: 0.5;
    vertical-align: middle;
}

.listings-table :deep(tbody tr.listing-row) {
    transition: background-color 120ms ease;
}

.listings-table :deep(tbody tr.listing-row:hover) {
    background-color: rgba(var(--v-theme-on-surface), 0.035);
}

.address-link {
    color: inherit;
    text-decoration: none;
    display: inline-flex;
    align-items: center;
    gap: 6px;
}

.address-link__icon {
    opacity: 0;
    transform: translateX(-4px);
    transition: opacity 120ms ease, transform 120ms ease;
    color: rgb(var(--v-theme-secondary));
}

.listing-row:hover .address-link__icon {
    opacity: 0.75;
    transform: translateX(0);
}

.address-link:hover {
    color: rgb(var(--v-theme-secondary));
    text-decoration: underline;
}

.address-link:hover .address-link__icon {
    opacity: 1;
}

.first-seen {
    display: inline-flex;
    align-items: center;
    gap: 8px;
}

.commute-link {
    color: rgb(var(--v-theme-secondary));
    text-decoration: none;
    display: inline-flex;
    align-items: center;
    gap: 6px;
    white-space: nowrap;
}

.commute-link__icon {
    opacity: 0.7;
    transition: opacity 120ms ease;
}

.commute-link:hover {
    text-decoration: underline;
}

.commute-link:hover .commute-link__icon {
    opacity: 1;
}
</style>
