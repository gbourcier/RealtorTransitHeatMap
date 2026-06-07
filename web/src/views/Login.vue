<template>
  <v-main class="login-main">
    <div class="login-backdrop" aria-hidden="true">
      <svg viewBox="0 0 1440 900" preserveAspectRatio="xMidYMid slice">
        <rect width="1440" height="900" fill="#171A16" />
        <rect x="120" y="520" width="260" height="200" fill="#1E241A" />
        <rect x="1040" y="180" width="300" height="220" fill="#1E241A" />
        <rect x="560" y="640" width="220" height="160" fill="#1E241A" />
        <path d="M0 700 Q260 640 520 706 T1040 700 T1500 690 L1500 900 L0 900 Z" fill="#15211F" />
        <g stroke="#23271F" stroke-width="10">
          <path d="M-20 150 H1500 M-20 320 H1500 M-20 500 H1500 M-20 690 H1500" />
          <path d="M240 60 V900 M520 60 V900 M800 60 V900 M1080 60 V900 M1320 60 V900" />
        </g>
        <g stroke="#34331E" stroke-width="6" opacity=".7">
          <path d="M-20 410 H1500" />
          <path d="M660 60 V900" />
        </g>
        <g>
          <circle cx="720" cy="430" r="230" fill="#9BE84A" opacity="0.10" />
          <circle cx="720" cy="430" r="150" fill="#FFB454" opacity="0.10" />
          <circle cx="720" cy="430" r="74" fill="#FF5C8A" opacity="0.14" />
        </g>
      </svg>
    </div>

    <section class="login-card" aria-labelledby="login-heading">
      <div class="brand-lockup">
        <div class="brand-mark">
          <img src="/brand/sona-logo.svg" alt="" width="42" height="42" />
        </div>
        <div class="wordmark" aria-label="HouseMap">
          House<span>Map</span>
        </div>
      </div>

      <h1 id="login-heading" class="heading">Welcome back</h1>
      <p class="subheading">Sign in to your commute heatmap</p>

      <form autocomplete="on" novalidate @submit.prevent="onSubmit">
        <div class="form-field">
          <label for="username">Username</label>
          <div class="input-shell input-shell--password" :class="{ 'input-shell--error': error }">
            <v-icon icon="mdi-account-outline" size="19" class="input-icon" />
            <input
              id="username"
              v-model="username"
              type="text"
              name="username"
              placeholder="username"
              autocomplete="username"
              required
              :disabled="loading"
            >
          </div>
        </div>

        <div class="form-field">
          <label for="password">Password</label>
          <div class="input-shell" :class="{ 'input-shell--error': error }">
            <v-icon icon="mdi-lock-outline" size="19" class="input-icon" />
            <input
              id="password"
              v-model="password"
              :type="showPassword ? 'text' : 'password'"
              name="password"
              placeholder="Password"
              autocomplete="current-password"
              required
              :disabled="loading"
            >
            <button
              type="button"
              class="reveal-button"
              :aria-label="showPassword ? 'Hide password' : 'Show password'"
              :disabled="loading"
              @click="showPassword = !showPassword"
            >
              <v-icon :icon="showPassword ? 'mdi-eye-off-outline' : 'mdi-eye-outline'" size="19" />
            </button>
          </div>
        </div>

        <p v-if="error" class="login-error" role="alert">{{ error }}</p>

        <button class="submit-button" type="submit" :disabled="loading || !username || !password">
          {{ loading ? 'Signing in...' : 'Sign in' }}
        </button>
      </form>

      <footer class="login-footer">
        <a href="https://github.com/gbourcier/RealtorTransitHeatMap" target="_blank" rel="noopener noreferrer">
          Source Code
        </a>
        <span aria-hidden="true"></span>
        <a href="mailto:housemap@gabriel.bourcier.app">Contact us</a>
      </footer>
    </section>
  </v-main>
</template>

<script lang="ts" setup>
import { shallowRef } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '../stores/auth'

const authStore = useAuthStore()
const router = useRouter()

const username = shallowRef('')
const password = shallowRef('')
const showPassword = shallowRef(false)
const loading = shallowRef(false)
const error = shallowRef('')

async function onSubmit() {
  if (!username.value || !password.value) return
  error.value = ''
  loading.value = true
  try {
    await authStore.login(username.value, password.value)
    router.push('/listings')
  } catch {
    error.value = 'Invalid username or password.'
    password.value = ''
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-main {
  --login-bg: #20231f;
  --login-card: #262925;
  --login-inset: #171a16;
  --login-ink: #f4f1e8;
  --login-muted: rgba(244, 241, 232, .52);
  --login-faint: rgba(244, 241, 232, .34);
  --login-line: rgba(244, 241, 232, .12);
  --login-line-hover: rgba(244, 241, 232, .22);
  --login-primary: #b6f24a;
  --login-primary-hover: #c6fb60;
  --login-on-primary: #172006;
  --login-secondary: #6ccff6;
  --login-error: #f92672;

  position: relative;
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 32px 20px;
  overflow: hidden;
  background: var(--login-bg);
  color: var(--login-ink);
  font-family: Inter, system-ui, sans-serif;
}

.login-backdrop {
  position: fixed;
  inset: 0;
  z-index: 0;
  pointer-events: none;
}

.login-backdrop svg {
  display: block;
  width: 100%;
  height: 100%;
}

.login-backdrop::after {
  position: absolute;
  inset: 0;
  content: "";
  background:
    radial-gradient(1100px 620px at 50% 42%, rgba(32, 35, 31, 0) 0%, rgba(28, 31, 27, .55) 58%, rgba(24, 26, 23, .92) 100%),
    radial-gradient(700px 360px at 84% -10%, rgba(38, 42, 34, .7) 0%, transparent 60%);
}

.login-card {
  position: relative;
  z-index: 1;
  width: min(100%, 432px);
  padding: 40px 38px 34px;
  background: var(--login-card);
  border: 1px solid var(--login-line);
  border-radius: 22px;
  box-shadow:
    0 40px 90px -30px #000,
    0 0 0 1px rgba(0, 0, 0, .4),
    inset 0 1px 0 rgba(255, 255, 255, .03);
}

.brand-lockup {
  display: flex;
  gap: 13px;
  align-items: center;
  justify-content: center;
  margin-bottom: 26px;
}

.brand-mark {
  display: grid;
  flex: 0 0 auto;
  width: 54px;
  height: 54px;
  place-items: center;
  background: linear-gradient(160deg, #fff, #eaf2fb);
  border-radius: 15px;
  box-shadow:
    inset 0 0 0 1px rgba(21, 35, 61, .1),
    0 4px 12px rgba(0, 0, 0, .4);
}

.brand-mark img {
  display: block;
}

.wordmark {
  font-family: "Baloo 2", Inter, system-ui, sans-serif;
  font-size: 30px;
  font-weight: 800;
  line-height: 1;
  color: var(--login-ink);
}

.wordmark span {
  color: var(--login-primary);
}

.heading {
  margin: 0 0 4px;
  font-size: 20px;
  font-weight: 700;
  line-height: 1.25;
  text-align: center;
}

.subheading {
  margin: 0 0 28px;
  font-size: 14px;
  font-weight: 500;
  line-height: 1.45;
  color: var(--login-muted);
  text-align: center;
}

form {
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.form-field label {
  display: block;
  margin: 0 0 8px 2px;
  font-size: 11px;
  font-weight: 600;
  line-height: 1;
  color: var(--login-faint);
  text-transform: uppercase;
  letter-spacing: .13em;
}

.input-shell {
  position: relative;
  display: flex;
  align-items: center;
}

.input-icon {
  position: absolute;
  left: 15px;
  color: var(--login-muted);
  pointer-events: none;
  transition: color .14s ease;
}

.input-shell input {
  width: 100%;
  height: 52px;
  padding: 0 16px 0 46px;
  font: inherit;
  font-size: 15px;
  font-weight: 500;
  color: var(--login-ink);
  outline: none;
  background: var(--login-inset);
  border: 1px solid var(--login-line);
  border-radius: 13px;
  transition: border-color .14s ease, box-shadow .14s ease, background .14s ease;
}

.input-shell--password input {
  padding-right: 54px;
}

.input-shell input::placeholder {
  font-weight: 400;
  color: var(--login-muted);
}

.input-shell input:hover {
  border-color: var(--login-line-hover);
}

.input-shell input:focus {
  background: #1b1f1a;
  border-color: var(--login-secondary);
  box-shadow: 0 0 0 3px rgba(108, 207, 246, .15);
}

.input-shell:focus-within .input-icon {
  color: var(--login-secondary);
}

.input-shell input:disabled {
  cursor: progress;
  opacity: .72;
}

.input-shell--error input,
.input-shell--error input:hover,
.input-shell--error input:focus {
  border-color: var(--login-error);
  box-shadow: 0 0 0 3px rgba(249, 38, 114, .14);
}

.reveal-button {
  position: absolute;
  right: 8px;
  display: grid;
  width: 38px;
  height: 38px;
  padding: 0;
  color: var(--login-muted);
  cursor: pointer;
  background: transparent;
  border: 0;
  border-radius: 9px;
  place-items: center;
  transition: background .14s ease, color .14s ease;
}

.reveal-button:hover,
.reveal-button:focus-visible {
  color: var(--login-ink);
  background: rgba(244, 241, 232, .07);
}

.reveal-button:focus-visible {
  outline: 2px solid var(--login-secondary);
  outline-offset: 2px;
}

.reveal-button:disabled {
  cursor: progress;
  opacity: .55;
}

.login-error {
  min-height: 20px;
  margin: 0;
  padding: 0 2px;
  font-size: 13px;
  font-weight: 600;
  line-height: 1.45;
  color: var(--login-error);
}

.submit-button {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 54px;
  margin-top: 2px;
  font: inherit;
  font-size: 16px;
  font-weight: 700;
  color: var(--login-on-primary);
  cursor: pointer;
  background: var(--login-primary);
  border: 0;
  border-radius: 14px;
  box-shadow: 0 2px 12px -2px rgba(182, 242, 74, .42);
  transition: background .14s ease, box-shadow .14s ease, transform .06s ease, opacity .14s ease;
}

.submit-button:hover:not(:disabled) {
  background: var(--login-primary-hover);
  box-shadow: 0 5px 18px -3px rgba(182, 242, 74, .55);
}

.submit-button:active:not(:disabled) {
  transform: translateY(1px);
}

.submit-button:focus-visible {
  outline: 2px solid var(--login-secondary);
  outline-offset: 3px;
}

.submit-button:disabled {
  cursor: default;
  opacity: .85;
}

.login-footer {
  display: flex;
  gap: 10px;
  align-items: center;
  justify-content: center;
  padding-top: 18px;
  margin-top: 24px;
  font-size: 12px;
  font-weight: 600;
  line-height: 1.4;
  color: var(--login-faint);
  border-top: 1px solid var(--login-line);
}

.login-footer a {
  color: var(--login-muted);
  text-decoration: none;
  transition: color .14s ease;
}

.login-footer a:hover,
.login-footer a:focus-visible {
  color: var(--login-secondary);
  text-decoration: underline;
}

.login-footer span {
  width: 5px;
  height: 5px;
  background: var(--login-primary);
  border-radius: 999px;
  opacity: .7;
}

@media (max-width: 520px) {
  .login-main {
    padding: 24px 14px;
  }

  .login-card {
    padding: 34px 22px 28px;
    border-radius: 20px;
  }

  .wordmark {
    font-size: 28px;
  }
}
</style>
