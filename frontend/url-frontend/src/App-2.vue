<script>
  export default {
    data() {
      return {
        longUrl: '',
        shortUrl: '',
        status: 'pending'
      };
    },
    methods: {
      toggleStatus() {
        if (this.status === 'active') {
          this.status = 'pending';
        } else if (this.status === 'pending') {
          this.status = 'inactive';
        } else {
          this.status = 'active';
        }
      }
    }
};
</script>

<template>
  <h1>Shorten Your URL</h1>
  <!-- <button v-on:click="toggleStatus">Change Status </button> -->
   <!-- @click is the same as v-on:click -->
  <button @click="toggleStatus">Change Status </button>
</template>

<style scoped>

</style>