<script setup>

import Navbar from '@/components/Navbar.vue';
import UrlForm from '@/components/UrlForm.vue';

// composition api 
const openNav = () => {
    var sidebar = document.getElementById('sidebar');

    if (sidebar.classList.contains("close")) {
        sidebar.classList.remove("close");
    } else {
        sidebar.classList.add("close");}
}
</script>

<template>
  <Navbar />
  <!-- Main Content -->
    <div class="content">
  <!-- Navbar -->
      <nav>
        <a role="button" @click="openNav">
          <i class="bx bx-menu"></i>
        </a>
      </nav>
  <!-- End of nav -->
   <UrlForm />

  </div>
</template>

<style scoped>

</style>