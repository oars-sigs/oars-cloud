<template>
  <div class="monitor">
      <iframe
      id="bdIframe"
      ref="iframe"
      frameborder="0"
      style="position:absolute;left: -60px"
      src="/proxy/system/grafana/http/3000/?search=open"
    ></iframe>
  </div>
</template>

<script>
export default {
  data() {
    return {
    };
  },
  mounted() {
      const oIframe = document.getElementById("bdIframe");
      const height = window.innerHeight - 80;
      const width = window.innerWidth - 190;
      oIframe.style.width =width + "px"; //数字是页面布局宽度差值
      oIframe.style.height = height + "px";
      console.log(width,height)
  },
  methods: {
  },
}
</script>

<style>
 .monitor {
     padding: 0 ;
     margin: 0 ;
 }
</style>