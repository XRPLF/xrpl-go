import type { SidebarsConfig } from "@docusaurus/plugin-content-docs";

// Group guides by reader task while keeping existing document IDs and URLs.
const sidebars: SidebarsConfig = {
  docsSidebar: [
    {
      type: "html",
      value: '<div class="sidebarSectionLabel">Introduction</div>',
      defaultStyle: true,
    },
    "intro",
    "installation",
    {
      type: "html",
      value: '<div class="sidebarSectionLabel">SDK guides</div>',
      defaultStyle: true,
    },
    {
      type: "category",
      label: "Clients and queries",
      items: [
        { type: "doc", id: "xrpl/rpc", label: "JSON-RPC" },
        { type: "doc", id: "xrpl/websocket", label: "WebSocket" },
        { type: "doc", id: "xrpl/queries", label: "Request and response types" },
      ],
    },
    {
      type: "category",
      label: "Wallets and transactions",
      items: [
        { type: "doc", id: "xrpl/wallet", label: "Wallets and signing" },
        { type: "doc", id: "xrpl/transaction", label: "Transactions" },
        { type: "doc", id: "xrpl/faucet", label: "Test funding" },
      ],
    },
    { type: "doc", id: "xrpl/ledger-entry-types", label: "Ledger data" },
    {
      type: "category",
      label: "Utilities",
      items: [
        { type: "doc", id: "xrpl/currency", label: "Currency amounts" },
        { type: "doc", id: "xrpl/flag", label: "Flags" },
        { type: "doc", id: "xrpl/hash", label: "Hashes" },
        { type: "doc", id: "xrpl/time", label: "Timestamps" },
      ],
    },
    {
      type: "html",
      value: '<div class="sidebarSectionLabel">Low-level packages</div>',
      defaultStyle: true,
    },
    "address-codec",
    "binary-codec",
    "keypairs",
    {
      type: "html",
      value: '<div class="sidebarSectionLabel">Confidential</div>',
      defaultStyle: true,
    },
    { type: "doc", id: "confidential/index", label: "Overview" },
    { type: "doc", id: "confidential/installation", label: "Installation" },
    { type: "doc", id: "confidential/builders", label: "Builders" },
    "confidential/mptcrypto",
    {
      type: "html",
      value: '<div class="sidebarSectionLabel">Migration</div>',
      defaultStyle: true,
    },
    "upgrading-from-v0.1.x-to-v0.2.0",
    "upgrading-from-v0.2.0-to-v0.3.0",
  ],
};

export default sidebars;
