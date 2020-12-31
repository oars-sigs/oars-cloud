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
      const height = window.innerHeight - 170;
      const width = window.innerWidth - 256;
      let cols = parseInt(width / 9, 10);
      let hostname = this.$route.query.hostname;
      let id = this.$route.query.id;
      this.navs.push({ text: this.$route.query.name });
      let term = new Terminal({
        cursorBlink: true,
        //rows: parseInt(height/20, 10),
        cols: cols,
      });
      this.term = term;
      let terminalContainer = document.getElementById("terminal");
      terminalContainer.style.height = height + "px";
      const fitAddon = new FitAddon();
      this.term.loadAddon(fitAddon);
      this.term.open(terminalContainer);
      fitAddon.fit();
      this.term.focus();
      this.term.scrollToBottom();
      let protocol = window.location.protocol == "https:" ? "wss" : "ws";
      let ws = new WebSocket(
        `${protocol}://${window.location.host}/api/exec/${hostname}/${id}`
      );
      this.ws = ws;
      this.ws.binaryType = "arraybuffer";
      this.ws.onmessage = function (e) {
        let buf = new TextDecoder().decode(e.data);
        term.write(buf);
      };
      term.onData((data) => {
        this.ws.send(data);
      });
      this.term._initialized = true;
    },
  },
};
</script>

<style>
</style>