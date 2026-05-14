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
const sortBy = ref<SortBy>("commute_time");
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

function formatDate(unix: number): string {
    return new Date(unix * 1000).toLocaleDateString("en-CA");
}

function formatCommute(seconds: number | null): string {
    if (seconds == null) return "—";
    return `${Math.round(seconds / 60)} min`;
}

onMounted(load);
</script>

<template>
    <v-container fluid class="pa-6">
        <v-card>
            <v-card-title class="d-flex align-center">
                <span>Listings</span>
                <v-spacer />
                <span
                    v-if="total > 0"
                    class="text-body-2 text-medium-emphasis mr-4"
                    >{{ total }} total</span
                >
                <v-btn
                    variant="text"
                    icon="mdi-refresh"
                    :loading="loading"
                    @click="load"
                />
            </v-card-title>
            <v-divider />

            <v-alert v-if="error" type="error" variant="tonal" class="ma-3">{{
                error
            }}</v-alert>

            <v-card-text
                v-if="loading && items.length === 0"
                class="text-center py-8"
            >
                <v-progress-circular indeterminate />
            </v-card-text>

            <v-table v-else-if="items.length > 0" density="comfortable">
                <thead>
                    <tr>
                        <th class="link-col" />
                        <th>Address</th>
                        <th
                            class="sortable-col text-right"
                            @click="toggleSort('price')"
                        >
                            Price
                            <v-icon size="small" class="sort-icon">{{
                                sortIcon("price")
                            }}</v-icon>
                        </th>
                        <th
                            class="sortable-col"
                            @click="toggleSort('first_seen_at')"
                        >
                            First Seen
                            <v-icon size="small" class="sort-icon">{{
                                sortIcon("first_seen_at")
                            }}</v-icon>
                        </th>
                        <th
                            class="sortable-col"
                            @click="toggleSort('commute_time')"
                        >
                            Commute Time
                            <v-icon size="small" class="sort-icon">{{
                                sortIcon("commute_time")
                            }}</v-icon>
                        </th>
                    </tr>
                </thead>
                <tbody>
                    <tr
                        v-for="item in items"
                        :key="`${item.board}-${item.mls}`"
                    >
                        <td class="link-col">
                            <v-btn
                                v-if="item.slug"
                                :href="item.slug"
                                target="_blank"
                                rel="noopener noreferrer"
                                variant="text"
                                icon="mdi-open-in-new"
                                size="small"
                                density="compact"
                            />
                        </td>
                        <td>{{ item.address || "—" }}</td>
                        <td class="text-right">
                            {{ formatPrice(item.currentPrice) }}
                        </td>
                        <td>{{ formatDate(item.firstSeenAt) }}</td>
                        <td>{{ formatCommute(item.commuteSecondsDowntown) }}</td>
                    </tr>
                </tbody>
            </v-table>

            <v-card-text v-else class="text-medium-emphasis text-center py-8">
                No listings found.
            </v-card-text>

            <template v-if="pageCount > 1">
                <v-divider />
                <v-card-text class="d-flex justify-center pa-3">
                    <v-pagination
                        :model-value="page"
                        :length="pageCount"
                        density="comfortable"
                        @update:model-value="
                            (p) => {
                                page = p;
                                load();
                            }
                        "
                    />
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
.link-col {
    width: 40px;
    padding-right: 0 !important;
}
</style>
