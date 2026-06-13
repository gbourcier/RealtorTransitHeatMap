import { onActivated, onBeforeUnmount, onDeactivated, onMounted } from "vue";

export function useBodyScrollLock(): void {
    let prevHtmlOverflow = "";
    let prevBodyOverflow = "";
    let locked = false;

    function lock(): void {
        if (locked) return;
        prevHtmlOverflow = document.documentElement.style.overflow;
        prevBodyOverflow = document.body.style.overflow;
        document.documentElement.style.overflow = "hidden";
        document.body.style.overflow = "hidden";
        locked = true;
    }

    function unlock(): void {
        if (!locked) return;
        document.documentElement.style.overflow = prevHtmlOverflow;
        document.body.style.overflow = prevBodyOverflow;
        locked = false;
    }

    onMounted(lock);
    onActivated(lock);
    onDeactivated(unlock);
    onBeforeUnmount(unlock);
}
