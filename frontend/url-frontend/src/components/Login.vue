<script setup>
import { ref } from 'vue';
import { useAuthStore } from '@/stores/auth';
import { storeToRefs } from 'pinia';
import { useRouter } from 'vue-router';
const router = useRouter();

const authStore = useAuthStore();
const { token, loading, error} = storeToRefs(authStore);
const { login, logout } = authStore;
const username = ref('');
const password = ref('');

const handleLogin = async () => {
   await login(username.value, password.value);

    if (token.value) {
         router.push('/');
    }
}
</script>

<template>
    <h1>Login</h1>
    <form @submit.prevent="handleLogin">
        <div>
            <label for="username">Username</label>
            <input type="text" id="username" v-model="username" name="username" />
        </div>
        <div>
            <label for="password">Password</label>
            <input type="password" id="password" v-model="password" name="password" />
        </div>
        <div>
            <button type="submit" :disabled="loading">
                {{ loading ? 'Logging in...' : 'Login' }}
            </button>
        </div>
        <div v-if="error" class="error">{{ error }}</div>
    </form>
</template>

<style scoped>
    .error {
        color: red;
    }
</style>

