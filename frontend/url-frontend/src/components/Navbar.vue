<script setup> </script>

<template>
<div class="sidebar" id="sidebar">
  <div class="side-nav">
    <a href="#" class="logo">
    <i class="bx bx-code-alt"></i>
    <div class="logo-name"><span>ALF</span></div>
    </a>
    <ul class="side-menu">
      <li class="active"><a href="/"><i class='bx bxs-dashboard'></i>Dashboard</a></li>
      <!-- <li><a href=""><i class='bx bxs-component'></i>VM</a></li>
      <li><a href="#"><i class='bx bxs-flask' ></i>Labs</a></li> -->
    </ul>
  </div>
</div>
<!-- End of sidebar -->
</template>

<style scoped>
.content nav {
    height: 56px;
    padding: 0 24px 0 0;
    display: flex;
    align-items: center;
    grid-gap: 24px;
    position: sticky;
    top: 0;
    left: 0;
    z-index: 1000;
    background: var(--dark-900-intermediary);
    border-width: 3px;
    border-style: solid;
    border-left: 0;
    border-top: 0;
    border-right: 0;
    border-image: linear-gradient(90deg, var(--primary-500) 0%, var(--secondary-500) 100%) 1;
}

.content nav a {
    color: var(--neutral-50);
    font-size: 1.2rem;
    cursor: pointer;
    padding-left: 10px;
}

.content nav a:hover {
    color: var(--primary-200);
}

.content nav .nav-btn {
    justify-content: flex-end;
    margin-right: auto;
    width: 100%;
    position: relative;
    display: flex;
    align-items: center;
}

.content nav .nav-btn a:nth-child(1) {
    padding-right: 20px;
}
</style>