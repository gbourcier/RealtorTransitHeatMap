import { onBeforeUnmount, onMounted, watch, type WatchSource } from "vue";

export const MOBILE_BOTTOM_SHEET_OPEN_EVENT = "mobile-bottom-sheet:open";

let nextSheetId = 0;

type Options = {
    open: WatchSource<boolean>;
    close: () => void;
    enabled?: () => boolean;
};

export function useMobileBottomSheetCoordinator(options: Options): void {
    const id = `mobile-bottom-sheet-${++nextSheetId}`;
    const enabled = options.enabled ?? (() => true);

    function isOpen(): boolean {
        return typeof options.open === "function"
            ? options.open()
            : options.open.value;
    }

    function announceOpen(): void {
        if (!enabled() || !isOpen()) return;
        window.dispatchEvent(
            new CustomEvent(MOBILE_BOTTOM_SHEET_OPEN_EVENT, {
                detail: { id },
            }),
        );
    }

    function onSheetOpen(event: Event): void {
        if (!(event instanceof CustomEvent)) return;
        if (event.detail?.id === id) return;
        if (!enabled() || !isOpen()) return;
        options.close();
    }

    watch(options.open, (open) => {
        if (open) announceOpen();
    }, { flush: "sync" });

    onMounted(() => {
        window.addEventListener(MOBILE_BOTTOM_SHEET_OPEN_EVENT, onSheetOpen);
    });

    onBeforeUnmount(() => {
        window.removeEventListener(MOBILE_BOTTOM_SHEET_OPEN_EVENT, onSheetOpen);
    });
}
