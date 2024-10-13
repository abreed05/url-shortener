<script setup>
import { jwtDecode } from "jwt-decode";
import { useAuthStore } from '@/stores/auth';
import { storeToRefs } from 'pinia';
import { useRouter } from 'vue-router';
import { ref } from 'vue';
import UrlForm from '@/components/UrlForm.vue';
const router = useRouter();
const authStore = useAuthStore();
const { token } = storeToRefs(authStore);


if (!token) {
    router.push('/login')
}

const decoded = jwtDecode(token.value);

if (decoded.exp < Date.now() / 1000) {
    authStore.logout();
    router.push('/login');
}

const username = decoded.sub;
const userId = ref(decoded.id)

</script>

<template>
    <h1>Home</h1>
    <p v-if="username">Welcome, {{ username }}</p>
    <p v-else>Not logged in</p>

    <UrlForm :userId="userId" />
</template>