<template>
  <v-container fluid class="px-0">
    <v-breadcrumbs :items="navs">
      <template v-slot:item="{ item }">
        <v-breadcrumbs-item :to="item.href" :disabled="item.disabled">
          {{ item.text.toUpperCase() }}
        </v-breadcrumbs-item>
      </template>
    </v-breadcrumbs>
    <div id="terminal" class="xterm"></div>
  </v-container>
</template>

<script>
import { Terminal } from "xterm";
import { FitAddon } from "xterm-addon-fit";
import "xterm/css/xterm.css";
export default {
  data() {
    return {
      log: "",
      navs: [
        {
          text: "端点",
          href: "/endpoint",
        },
        {
          text: "终端",
        },
      ],
      ws: null,
      term: null,
    };
  },
  mounted() {
    this.init();
  },
  beforeDestroy() {
    this.ws.close();
    this.term.destroy();
  },
  methods: {
    init: function () {
      let hostname= this.$route.query.hostname;
      let id= this.$route.query.id;
      this.navs.push({text:hostname})
      this.navs.push({text:this.$route.query.name})
      let term = new Terminal();
      this.term=term
      let terminalContainer = document.getElementById("terminal");
      const fitAddon = new FitAddon();
      this.term.loadAddon(fitAddon);
      this.term.open(terminalContainer);
      fitAddon.fit();
      this.term.focus();
      let ws = new WebSocket(`ws://${window.location.host}/api/exec/${hostname}/${id}` );
      this.ws=ws
      this.ws.binaryType = "arraybuffer";
      this.ws.onmessage = function (e) {
          let buf = new TextDecoder().decode(e.data);
          term.write(buf);
      };
        term.onData(data=>{
            this.ws.send(data);
        });
      this.term._initialized = true;
    },
  },
};
</script>

<style>
</style>