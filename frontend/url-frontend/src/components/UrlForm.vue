<script setup>
import { ref, toRefs  } from 'vue';
import { useAuthStore } from '@/stores/auth';
import { storeToRefs } from 'pinia';
const longUrl = ref('');
import { useRouter } from 'vue-router';
const router = useRouter();
const authStore = useAuthStore();
const { token } = storeToRefs(authStore);
const props = defineProps({
  userId: String
})

const shortenUrl = async () => {
  if (longUrl.value.trim() !== '') {
    try {
      const response = await fetch(import.meta.env.VITE_API_URL, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': token.value
        },
        body: JSON.stringify({
          long_url: longUrl.value,
          user_id: props.userId
        })
      });
      const data = await response.json();
      
    } catch (error) {
      console.error('Error:', error);
    }
  }
}
</script>

<template>
  <main class="parent">
    <div class="header">
      <h1>Shorten Your URL</h1>
    </div>
    <div>
      <form @submit.prevent="shortenUrl" class="url-form">
        <div class="input-group">
          <label for="longUrl">Enter your URL</label>
          <input type="text" id="longUrl" v-model="longUrl" name="long_url" />
        </div>
        <div class="url-submit">
          <button type="submit" class="url-btn">Submit</button>
        </div>
      </form>
    </div>
  </main>
</template>

<style scoped>
  .parent {
    display: flex;
    align-items: center;
    justify-content: center;
    flex-direction: column;
    margin-top: 20px;
    color: var(--neutral-200);
  }

  .header {
    display: flex;
    align-items: center;
    justify-content: space-between;
    padding: 0 24px;
    flex-direction: column;
  }

  .url-form .input-group {
    display: flex;
    flex-direction: column;
    align-items: flex-start;
  }

  .url-form .input-group label {
    color: var(--neutral-50);
    font-size: 1rem;
    padding: 5px 0;
    margin-top: 10px;
  }

  .url-form .input-group input {
    flex-grow: 1;
    padding: 0 16px;
    border: 1px solid rgba(255,255,255,0.2);
    background: var(--dark-900-intermediary);
    color: var(--neutral-200);
    border-radius: 10px;
    outline: none;
    width: 400px;
    height: 35px;
    margin-top: 10px;
    box-sizing: content-box;
  }

  .url-form .url-submit {
    display: flex;
    justify-content: flex-end;
    margin-top: 10px; 
    gap: 10px;
}

.url-form .url-submit .url-btn {
    background: var(--purple);
    color: white;
    padding: 8px 16px;
    border-radius: 10px;
    font-size: 16px;
    cursor: pointer;
}
</style>