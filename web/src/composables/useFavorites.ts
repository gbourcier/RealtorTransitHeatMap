import { ref, type InjectionKey, type Ref } from "vue";
import { addFavorite, removeFavorite } from "../api/favorites";

const UNDO_TIMEOUT_MS = 5000;

export interface FavoriteToggleItem {
    board: number;
    mls: number;
    isFavorite: boolean;
}

export interface FavoritesSnackbarState {
    open: boolean;
    count: number;
}

export interface FavoritesController {
    isFavorite: (board: number, mls: number, serverValue: boolean) => boolean;
    toggle: (item: FavoriteToggleItem) => void;
    flush: () => void;
    undo: () => void;
    snackbar: Ref<FavoritesSnackbarState>;
    onError: (cb: (message: string) => void) => void;
}

export const favoritesKey: InjectionKey<FavoritesController> =
    Symbol("favorites");

type PendingRemoval = {
    item: FavoriteToggleItem;
    timer: ReturnType<typeof setTimeout>;
};

function keyOf(board: number, mls: number): string {
    return `${board}-${mls}`;
}

export function useFavorites(): FavoritesController {
    const overrides = ref<Record<string, boolean>>({});
    const pending = new Map<string, PendingRemoval>();
    const snackbar = ref<FavoritesSnackbarState>({ open: false, count: 0 });
    let errorCb: ((message: string) => void) | null = null;

    function setOverride(k: string, value: boolean): void {
        overrides.value = { ...overrides.value, [k]: value };
    }

    function refreshSnackbar(): void {
        snackbar.value = { open: pending.size > 0, count: pending.size };
    }

    function isFavorite(
        board: number,
        mls: number,
        serverValue: boolean,
    ): boolean {
        const o = overrides.value[keyOf(board, mls)];
        return o === undefined ? serverValue : o;
    }

    async function commitAdd(item: FavoriteToggleItem): Promise<void> {
        try {
            await addFavorite(item.board, item.mls);
        } catch (e: any) {
            setOverride(keyOf(item.board, item.mls), false);
            errorCb?.(e?.response?.data?.error ?? "Failed to add favorite.");
        }
    }

    async function commitRemove(item: FavoriteToggleItem): Promise<void> {
        try {
            await removeFavorite(item.board, item.mls);
        } catch (e: any) {
            setOverride(keyOf(item.board, item.mls), true);
            errorCb?.(e?.response?.data?.error ?? "Failed to remove favorite.");
        }
    }

    function toggle(item: FavoriteToggleItem): void {
        const k = keyOf(item.board, item.mls);
        const next = !isFavorite(item.board, item.mls, item.isFavorite);
        setOverride(k, next);

        if (next) {
            const existing = pending.get(k);
            if (existing) {
                clearTimeout(existing.timer);
                pending.delete(k);
                refreshSnackbar();
                return;
            }
            void commitAdd(item);
            return;
        }

        const existing = pending.get(k);
        if (existing) clearTimeout(existing.timer);
        const pendingRemoval: PendingRemoval = {
            item,
            timer: setTimeout(() => {
                if (pending.get(k) !== pendingRemoval) return;
                pending.delete(k);
                refreshSnackbar();
                void commitRemove(item);
            }, UNDO_TIMEOUT_MS),
        };
        pending.set(k, pendingRemoval);
        refreshSnackbar();
    }

    function undo(): void {
        for (const [k, p] of pending) {
            clearTimeout(p.timer);
            setOverride(k, true);
        }
        pending.clear();
        refreshSnackbar();
    }

    function flush(): void {
        for (const [k, p] of pending) {
            clearTimeout(p.timer);
            pending.delete(k);
            void commitRemove(p.item);
        }
        refreshSnackbar();
    }

    function onError(cb: (message: string) => void): void {
        errorCb = cb;
    }

    return { isFavorite, toggle, flush, undo, snackbar, onError };
}
