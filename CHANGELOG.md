# Changelog

All notable changes to this project will be documented in this file.

## [0.67.0](https://github.com/kowabunga-cloud/kahuna/compare/v0.66.0...v0.67.0) (2026-10-11)

### Features

* add health endpoints and version information ([0023a32](https://github.com/kowabunga-cloud/kahuna/commit/0023a3246297bac7d83540b4fda09934362627e2))

## [0.66.0](https://github.com/kowabunga-cloud/kahuna/compare/v0.65.0...v0.66.0) (2026-10-11)

### Features

* update to API v0.58.0 ([c45fb06](https://github.com/kowabunga-cloud/kahuna/commit/c45fb0692c67cc57d8d789be3ae969d2d0568abf))

## [0.65.0](https://github.com/kowabunga-cloud/kahuna/compare/v0.64.2...v0.65.0) (2026-10-11)

### Features

* add conditional uefi support to instances and komputes ([3dfba54](https://github.com/kowabunga-cloud/kahuna/commit/3dfba5472d3c1505f1ed2ee76369f83c62334127))
* add kmotion_enabled field, enabled unless explicitly disabled ([b0414c6](https://github.com/kowabunga-cloud/kahuna/commit/b0414c66b2344926480f6fc55ce229408f9e8a58))
* add secure boot tpm and hyper-v enlightenments for windows 11 ([8c09ce0](https://github.com/kowabunga-cloud/kahuna/commit/8c09ce0264a767d4da06ffbaed183f4dbec3695d))
* bump SDK to v0.56.0 and stub kMotion/Kwarantine boilerplate ([d5fea56](https://github.com/kowabunga-cloud/kahuna/commit/d5fea566b17ab71a849f41bf8b4ab02550f83a9f))
* enforce uefi for windows and optimize generic q35 topology ([20bf830](https://github.com/kowabunga-cloud/kahuna/commit/20bf830249f0fcb3205e02051ceaff42aa5950f3))
* implement cold kMotion commit with rollback ([c606a3d](https://github.com/kowabunga-cloud/kahuna/commit/c606a3d72ee12dbfc13e59136c64a28f4aeac5dd))
* implement kMotion plan computation ([e9a7359](https://github.com/kowabunga-cloud/kahuna/commit/e9a7359a54c408ae69989986e61822b83f4f83ee))
* implement Kwarantine anti-affinity groups and scheduling exclusion ([908f30b](https://github.com/kowabunga-cloud/kahuna/commit/908f30b40517f3b58ae8b788023509b086e32499))
* implement live kMotion commit via peer-to-peer migration ([aac1b84](https://github.com/kowabunga-cloud/kahuna/commit/aac1b84a32fe2e0742cc9457179f6891644bd1f0))
* implement region-scoped instance creation and listing ([be1df63](https://github.com/kowabunga-cloud/kahuna/commit/be1df630b0659443fababd099cfb01861bcdd33c))
* implement region-scoped kompute creation and listing ([65c936f](https://github.com/kowabunga-cloud/kahuna/commit/65c936fef4ea303008112cf4ce75e38b2c58dadf))
* optimize libvirt domain XML for near-native perf and safe migration ([0ad39b7](https://github.com/kowabunga-cloud/kahuna/commit/0ad39b779b5a5233f739cf469de4713f2679bf1a))
* upgrade machine architecture to q35 with sata cdrom ([27eac45](https://github.com/kowabunga-cloud/kahuna/commit/27eac45a699dbd4f01e763739489d7986c67aa8f))
* wire kwarantines field into instance/kompute creation ([6b6a41b](https://github.com/kowabunga-cloud/kahuna/commit/6b6a41b653662b1f5961fb614a66fd1368f26700))

### Bug Fixes

* actually delete adapters when rolling back a failed Kompute ([58dc0ac](https://github.com/kowabunga-cloud/kahuna/commit/58dc0ac877efac6021d896e78486969e87acbe1e))
* allocate IP addresses, VRRP IDs and private subnets atomically ([8a6121c](https://github.com/kowabunga-cloud/kahuna/commit/8a6121cd4cb226c8fc300efb73ca1eab3e8e3eb7))
* always remove cloud-init temporary files ([fd082ea](https://github.com/kowabunga-cloud/kahuna/commit/fd082ea10d9652d2f0be016bfd5ea44c9524e2b5))
* authorize requests on the anchored URL path only ([d9c472d](https://github.com/kowabunga-cloud/kahuna/commit/d9c472d80bfcce2dabd5be834db158aedc47afa9))
* bound and revoke the API key authentication cache ([1c74f8b](https://github.com/kowabunga-cloud/kahuna/commit/1c74f8b3513f756a27260db867c849cc85eb9958))
* bound HTTP request header read time and body size ([bd1d88d](https://github.com/kowabunga-cloud/kahuna/commit/bd1d88d4230534393da7adcbed8d86fa5d2f4c0c))
* bound MongoDB operations with a default timeout ([5558b47](https://github.com/kowabunga-cloud/kahuna/commit/5558b477450d8e6687db3eb704994fee0924177f))
* don't corrupt zone Kaktus references during Konvey placement ([112f968](https://github.com/kowabunga-cloud/kahuna/commit/112f9687fa9ea5a6b025fa3598acabe77d8a688f))
* fail cloud-init generation when a template can't be rendered ([3233587](https://github.com/kowabunga-cloud/kahuna/commit/323358734f3c16be85bf83939e87e2973addb601))
* fail Kawaii creation cleanly in regions without public network ([7cf25b3](https://github.com/kowabunga-cloud/kahuna/commit/7cf25b3a4e4bb96873513a62ca575605d8df0405))
* fail libvirt domain definition when a disk or NIC can't be resolved ([2db0ac6](https://github.com/kowabunga-cloud/kahuna/commit/2db0ac60b04a51090ad95825d4f11ce506fca112))
* filter project zone and region listings by location ([a1e7802](https://github.com/kowabunga-cloud/kahuna/commit/a1e78023b6fd424903f8ac345263baf70f4488db))
* invalidate cache entries when deleting resources ([160f236](https://github.com/kowabunga-cloud/kahuna/commit/160f2368973ab96658bd5b7578d03390e615966b))
* invalidate cached user resources on ownership changes ([13f30d8](https://github.com/kowabunga-cloud/kahuna/commit/13f30d81c1265444cc873068d59795abb4eaa60d))
* leave omitted fields untouched on partial updates ([172f6ba](https://github.com/kowabunga-cloud/kahuna/commit/172f6ba681adaa4ef89d9f808ce764bdbd83c17c))
* look up storage pool templates by their actual field name ([0cd5b0c](https://github.com/kowabunga-cloud/kahuna/commit/0cd5b0ce5dd10fe3636ce14cb22fd708400dfe69))
* make agent disconnection and reconnection reliable ([69db8ba](https://github.com/kowabunga-cloud/kahuna/commit/69db8ba9c5ea2fa8999edb92f91a93991ded11c7))
* make resource deletions resumable after a partial failure ([4f212da](https://github.com/kowabunga-cloud/kahuna/commit/4f212da89b4a5932364d385291da2e08b9d838ab))
* never overwrite atomically managed fields when saving resources ([5eab919](https://github.com/kowabunga-cloud/kahuna/commit/5eab919ce9649ada7d6ae1ab950d854c54f8eadc))
* only attach own, unused volumes and adapters to instances ([3f4931d](https://github.com/kowabunga-cloud/kahuna/commit/3f4931daae1bfc4b0263ed5c43bc0a1f9cdb5dcc))
* only drop child references to resources that are actually gone ([342ecda](https://github.com/kowabunga-cloud/kahuna/commit/342ecda1ad0e6a2af8f7f947aa35e6057bd99cf9))
* only update the cost field when reading a project's cost ([b0011c9](https://github.com/kowabunga-cloud/kahuna/commit/b0011c980d1d6d9b23795f2afbc48c5098f9590a))
* query instance collection when finding instances by kaktus ([4608278](https://github.com/kowabunga-cloud/kahuna/commit/4608278bfd0d3bb2af89737b7eb28b3417a6489d))
* refuse to delete projects that still have any child resource ([b406dc6](https://github.com/kowabunga-cloud/kahuna/commit/b406dc6030a83a7245897a49baa2e7f686e1bc08))
* refuse to start with an empty JWT signature or lifetime ([15aaf41](https://github.com/kowabunga-cloud/kahuna/commit/15aaf41fd3090db755fb1feb230aa27c465b9dd6))
* reject empty user registration and password renewal tokens ([24087fc](https://github.com/kowabunga-cloud/kahuna/commit/24087fcbf9260413ee19e2bfab924d47aab30480))
* reject negative resource sizes to prevent quota bypass ([948b19d](https://github.com/kowabunga-cloud/kahuna/commit/948b19d21a4576462565dcc67d8e20e21c0b330a))
* reject non-Bearer Authorization headers instead of slicing them ([35c6fb0](https://github.com/kowabunga-cloud/kahuna/commit/35c6fb07446bfbf6b011bc1bcc2383bee3619358))
* reject other projects' subnets and Kwarantine groups in requests ([2feb0a4](https://github.com/kowabunga-cloud/kahuna/commit/2feb0a4bc495f734a0ea1dffe1c698e5247a6932))
* reject zone Konvey creation only when the name is already taken ([57d055b](https://github.com/kowabunga-cloud/kahuna/commit/57d055b7be82ee4e9bb862f0f68807c5905d4421))
* require agent credentials for agent-managed instances metadata ([7a0335e](https://github.com/kowabunga-cloud/kahuna/commit/7a0335e22ee5fc357b6232b00612ef6a706b5f6b))
* reserve and release private subnets on project regions update ([abde08a](https://github.com/kowabunga-cloud/kahuna/commit/abde08a7a2aa7101c246624d939d3dee2b0f6ef8))
* revoke JWT sessions on logout and password change ([12858cc](https://github.com/kowabunga-cloud/kahuna/commit/12858cc4311dc2e7fbe8f06c739bef6ebfb85de2))
* roll back partially failed resource creations ([33d564d](https://github.com/kowabunga-cloud/kahuna/commit/33d564df537290c45987822def981bf2fe1b876a))
* scope Kawaii IPsec connection lookups to their parent Kawaii ([489154c](https://github.com/kowabunga-cloud/kahuna/commit/489154c7de48bf29a6be6191c24cd474c380fbaa))
* stop killing Kahuna on systemd reload and early stop timeout ([8f100e7](https://github.com/kowabunga-cloud/kahuna/commit/8f100e7eebde8bd7c97144c7b0a3fbb9d754bfd0))
* stop metrics scrapes from overwriting projects and exporting stale data ([0e573b4](https://github.com/kowabunga-cloud/kahuna/commit/0e573b4a9acb64c58f3e61711f5cd070308cf915))
* stop reporting cache invalidation misses as save failures ([d086e7f](https://github.com/kowabunga-cloud/kahuna/commit/d086e7f83c39f31c3edabc9078bed95b9a0e91fe))
* update child references and usage counters atomically ([53f9d44](https://github.com/kowabunga-cloud/kahuna/commit/53f9d442087ad652ed123795fe781f4623ef31d3))
* validate values rendered into cloud-init user-data ([5e20eab](https://github.com/kowabunga-cloud/kahuna/commit/5e20eabf7c70acb968fa9806dace15016450293a))

### Performance Improvements

* align MongoDB indexes with actual query fields ([2563eb1](https://github.com/kowabunga-cloud/kahuna/commit/2563eb19131df780efe6ca59c871317b61658380))
* look up user API keys by indexed SHA-256 digest ([1eb293c](https://github.com/kowabunga-cloud/kahuna/commit/1eb293c618f446befaeb5c26dc6362e2a465a521))
* query DNS records by domain and name instead of scanning ([c0bbbf9](https://github.com/kowabunga-cloud/kahuna/commit/c0bbbf91f05ca5999c7a9e5fcc2d013de850e7b2))

### Documentation

* document how to run MongoDB integration tests ([0b6ae56](https://github.com/kowabunga-cloud/kahuna/commit/0b6ae56a4738781f9ae3a4e14afc6714d4127bcc))

### Chores

* add AGENTS.md symlink ([c806a41](https://github.com/kowabunga-cloud/kahuna/commit/c806a415867fc70409f8bb565797f61ed307d6f6))
* add GitHub Sponsors FUNDING.yml ([5ea333b](https://github.com/kowabunga-cloud/kahuna/commit/5ea333ba5b19ff95d9f128785ac66a1986c329a5))
* add MongoDB indexes and project IDs in list endpoints ([b71d022](https://github.com/kowabunga-cloud/kahuna/commit/b71d02286d9cefe50e8b08090e229ad7bb920d75))
* better way to install golangci-lint ([8bc21d8](https://github.com/kowabunga-cloud/kahuna/commit/8bc21d8c95d6627ee83e0388bbac41bf2a50a184))
* bump common dependency to v0.65.0 ([d209899](https://github.com/kowabunga-cloud/kahuna/commit/d20989912e811f22b6676b88a3180072f1db5d5e))
* bump common dependency to v0.66.0 ([9be4495](https://github.com/kowabunga-cloud/kahuna/commit/9be449545f423f88ed50bd9cf1c420a6c991d7f8))
* bump SDK to v0.57.0, extend scheduling for region-wide election ([3a854e7](https://github.com/kowabunga-cloud/kahuna/commit/3a854e74b347da42b68450c2c0d649144cf096ff))
* **deps:** bump golang.org/x/net to v0.60.0 ([ad91f05](https://github.com/kowabunga-cloud/kahuna/commit/ad91f057d2263ec2f4f5b562e38b6c14ceda176b))
* do not run gosec on pre-commit ([390c5de](https://github.com/kowabunga-cloud/kahuna/commit/390c5de9ce50ed3c5f4e49ac58ddb6cef929dc33))
* fix cache retrieval pointer semantics and reduce cache allocations ([99769a7](https://github.com/kowabunga-cloud/kahuna/commit/99769a7d227257c26709b28ff35930929d861764))
* fix concurrent map access on agents and optimize agent selection ([7fb3374](https://github.com/kowabunga-cloud/kahuna/commit/7fb3374ec323a5220d523bdf321778c5dce475dc))
* fix security warnings reported by gosec ([8de95d2](https://github.com/kowabunga-cloud/kahuna/commit/8de95d2a09c47e71311823dfb6a7cf16b3e53654))
* guard nil-pointer edge cases and optimize ipsec and token handling ([1832696](https://github.com/kowabunga-cloud/kahuna/commit/18326968ab848ecf42d9a4ebba6c26e745a04f1c))
* guard SDK reflection logging and pre-compile validation regexes ([7361900](https://github.com/kowabunga-cloud/kahuna/commit/7361900882d1af533b337665dd86a71353b4cf1b))
* ignore .agents folder ([f236984](https://github.com/kowabunga-cloud/kahuna/commit/f236984b31da9940cb4b8a56cfae3c89f4a8ea31))
* implement SDK v0.55.0; add support for Kaktus node maintenance ([cbc2024](https://github.com/kowabunga-cloud/kahuna/commit/cbc2024d7b24fd2cd7decada6dede856c3d95da0))
* improve router auth check by only querying user tokens ([5bfaad6](https://github.com/kowabunga-cloud/kahuna/commit/5bfaad6b8ca881317036e0703eea691062d8891a))
* introduce DB/cache decoupling ([7fe7a14](https://github.com/kowabunga-cloud/kahuna/commit/7fe7a14b6a8dcb865102c1a092158f92d0804f96))
* optimize adapter IP assignment with O(1) lookups ([d22b482](https://github.com/kowabunga-cloud/kahuna/commit/d22b48294448187ef6b5f2c40427f7c3b2dad3e6))
* optimize API token authentication and route authorization checks ([59b53de](https://github.com/kowabunga-cloud/kahuna/commit/59b53de7d4bbe6acfee904456aa0cc7b4b389a14))
* optimize child reference lookups, slice operations, and sort algorithms ([944db75](https://github.com/kowabunga-cloud/kahuna/commit/944db754c0b6ed3eccccf3fac9ef694921aec334))
* optimize IP arithmetic and subnet address count algorithms ([b62b74f](https://github.com/kowabunga-cloud/kahuna/commit/b62b74f2b4dda9dcfc119c2f0d28e3ef002b22a0))
* optimize metrics scraping queries and cloud-init file handling ([1849d4e](https://github.com/kowabunga-cloud/kahuna/commit/1849d4ed99535961b3cab2d571b6e7001882fe27))
* optimize singleton initializers and eliminate config lock contention ([ac7b3a5](https://github.com/kowabunga-cloud/kahuna/commit/ac7b3a5e86f96a881faf42da0c530e50f968df28))
* require go 1.27.1 ([24de2b8](https://github.com/kowabunga-cloud/kahuna/commit/24de2b88e1ea2f277dd3c5fab3dd7f12656f8336))
* simplified internal packages handling ([4923dce](https://github.com/kowabunga-cloud/kahuna/commit/4923dce552942d159c38ec817f2aa922960d2049))
* standardize IP calculations on netip and drop ipaddress-go dependency ([d937618](https://github.com/kowabunga-cloud/kahuna/commit/d937618ebb999ee45aa360e0013ef65702fe6251))
* synchronize HTTP graceful shutdown with server cleanup ([42ee44e](https://github.com/kowabunga-cloud/kahuna/commit/42ee44ee1b7a0e8bbf490b5c9481790617ea5c3b))
* update dependencies ([f63be52](https://github.com/kowabunga-cloud/kahuna/commit/f63be52125042aeee322ce3abb330d9fb748a5a3))
* update dependencies ([75a46d3](https://github.com/kowabunga-cloud/kahuna/commit/75a46d39d1d0d3437e96c026580b7708a87512c6))
* update dependencies ([fd46da7](https://github.com/kowabunga-cloud/kahuna/commit/fd46da795ecf69dbf00c103f0d69434478cf63ee))
* update openapi sdk to v0.54.0 ([058cced](https://github.com/kowabunga-cloud/kahuna/commit/058cced019f89bb99a4843b84014e2b07660e7e2))
* use constant time string compare on master api key check to prevent timing side-channels attacks ([718a623](https://github.com/kowabunga-cloud/kahuna/commit/718a623c62986973648376cc18279fecab4c755e))
* use pre-compiled regexp for router queries parsing ([4215023](https://github.com/kowabunga-cloud/kahuna/commit/42150236be20a5661b35eeb4eb01a6948bc7d316))

## [0.64.2](https://github.com/kowabunga-cloud/kahuna/compare/v0.64.1...v0.64.2) (2026-04-03)

### Chores

* add editor config ([b1b22aa](https://github.com/kowabunga-cloud/kahuna/commit/b1b22aa9395224db48cd883c4323a6aef8bdc4ee))
* add pre-commit hook ([4401dad](https://github.com/kowabunga-cloud/kahuna/commit/4401dad37fe37930d145d16ef32b5bf37a93e06e))
* add semantic-release support ([488e768](https://github.com/kowabunga-cloud/kahuna/commit/488e768cbee9fb36efe5e08afe67b7a523d36b67))
* always use latest version of build tools ([ed9643a](https://github.com/kowabunga-cloud/kahuna/commit/ed9643ab01c5893332b234f6c3fea9753b248b54))
* don't complain on gosec G101 error ([05b5a6f](https://github.com/kowabunga-cloud/kahuna/commit/05b5a6f7d5d6302fd96b0b0e9bcceefd0f304fd5))
* don't exclude any gosec codes by default ([b1e299c](https://github.com/kowabunga-cloud/kahuna/commit/b1e299c18cd3d27d59b3f56ac9d0e4d546429c6f))
* explicit gosec errors ignore as all are integer conversions between signed/unsigned types that are safe in context — memory/storage sizes (always positive), CPU/instance counts (always small), VRRP IDs (bounded 1–255), and libvirt XML field types that require uint ([3f84afe](https://github.com/kowabunga-cloud/kahuna/commit/3f84afead29d7d9a0f543ebd0a9c07fd75ea466f))
* fix Makefile directive name ([fc6abe4](https://github.com/kowabunga-cloud/kahuna/commit/fc6abe44a83db8075ab1b2cd4a42a3eff58cc836))
* fix pre-commit issues ([a021ba3](https://github.com/kowabunga-cloud/kahuna/commit/a021ba3de76d4636ecea32a546af8165cbc520bf))
* ignore more files ([088672b](https://github.com/kowabunga-cloud/kahuna/commit/088672bc56a73bb65de072ce46283b3782a3347a))
* implement more unit tests ([d02200a](https://github.com/kowabunga-cloud/kahuna/commit/d02200aaf49dc657bf03404cb95fad2a538317ac))
* intentionally don't run gosec on sdk auto-generated code (won't be fixed) ([dba7737](https://github.com/kowabunga-cloud/kahuna/commit/dba773759e4813df55b1abe3b5d70eb9774e7db9))
* re-use shared GH action workflows ([51ee64c](https://github.com/kowabunga-cloud/kahuna/commit/51ee64c57b41322cbbdadb6cca597be6402e1ef3))
* remove apk build ([eab03aa](https://github.com/kowabunga-cloud/kahuna/commit/eab03aae27dfe65a1d276e9af98d0eda7faa6240))
* remove useless version and codename flag ([c917ae6](https://github.com/kowabunga-cloud/kahuna/commit/c917ae6081d618d8eadd55b051d16753eae8b3f5))
* require go 1.26 ([f3b6d13](https://github.com/kowabunga-cloud/kahuna/commit/f3b6d137e0851e277c9b7bca390e0e10f02b4ba9))
* simplified debian packaging ([a1652a2](https://github.com/kowabunga-cloud/kahuna/commit/a1652a284858edb1b2b8969a725ef42756fff972))
* update dependencies ([bec9759](https://github.com/kowabunga-cloud/kahuna/commit/bec97593f86c6ef885a88c312fb9a23fbd514db4))

### Documentation

* add Claude doc ([801afef](https://github.com/kowabunga-cloud/kahuna/commit/801afef826506ec28fc97fd00c6a77107ae1d325))
* add code owners ([a736537](https://github.com/kowabunga-cloud/kahuna/commit/a736537ad5a014ed42604a149c530834890f445c))
* remove empty white line ([714ba9a](https://github.com/kowabunga-cloud/kahuna/commit/714ba9a356e7d559e83ffa600e208fb19b642d11))
* updated logo ([aa5d479](https://github.com/kowabunga-cloud/kahuna/commit/aa5d479cdfd04456e5f84954356a4043169c12ed))
* updated README ([8352668](https://github.com/kowabunga-cloud/kahuna/commit/835266829e072627b3fc7f07303d1313956a6a6c))

## unreleased (yyyy-mm-dd)

## 0.64.1 (2025-12-30)
* **NEW**: use external Kowabunga 'common' package.

## 0.64.0 (2025-12-30)
* **BREAKING CHANGE**: deprecate complete OSX build (failed Ceph libraries dependency).
* **NEW**: implemented public API **v0.53.2**.
* **NEW**: **kiwi**: deprecate PowerDNS support in favor of built-in DNS server.
* **NEW**: **kiwi**: implement Reload() RPC call.
* **NEW**: implement proper common kahuna <-> agents protocol definition package.
* **NEW**: Update build dependencies.

## 0.63.3 (2025-09-11)
* **BUG**: **kompute**: fix some private IP assignment issue when 'public' (exposed) network contains private subnets.

## 0.63.2 (2025-09-03)
* **NEW**: Add support for ARM64 architecture.
* **NEW**: Update build dependencies.
* **NEW**: updated gosec to v2.22.8.
* **NEW**: updated golangci-lint to v2.4.0.
* **BUG**: correct SMTP email format (html first, plain text as fallback)

## 0.63.1 (2025-05-08)
* **NEW**: updated logo in email notifications.
* **NEW**: updated dependencies.
* **BUG**: fix APT packages repo URL in Linux cloud-init.

## 0.63.0 (2025-05-02)
* **NEW**: **kahuna**: switched to MongoDB driver v2.
* **NEW**: **kawaii**: When creating a Kawaii Public and Private VIP, those are now coupled under a same Virtual Router ID
* **NEW**: **misc**: upgraded to golangci-lint v2, fixes compliance issues.
* **BUG**: **kawaii**: IPsecs routing is updated dynamically on VIP failover
* **BUG**: **kawaii**: add a firewall rule to allow AH and ESP protocols  in a tunnel

## 0.62.6 (2025-03-06)
* **NEW**: requires Go SDK 1.24.

## 0.62.5 (2025-03-06)
* **NEW**: implemented public API **v0.52.3**.

## 0.62.4 (2025-03-05)
* **NEW**: updated dependencies.
* **NEW**: updated gosec to v2.22.2.
* **NEW**: updated govulncheck to v1.1.4.
* **NEW**: updated golangci-lint to v1.64.6.
* **BUG**: **kaktus**: fix segmentation fault if downloaded QCOW image does not have additional headers field.
* **BUG**: **kahuna**: check for template name unicity per pool, not globally.

## 0.62.3 (2025-02-20)
* **NEW**: updated dependencies.
* **BUG**: **kawaii**: add forward rules on firewall to allow traffic between peered subnets

## 0.62.2 (2025-01-14)
* **NEW**: updated dependencies.
* **BUG**: **kahuna**: extend router permissions for non-admin users to query **region** and **zone** endpoints.

## 0.62.1 (2024-12-17)
* **NEW**: updated dependencies.
* **BUG**: **kawaii**: enforce proper OpenSWAN service reload at configuration change.

## 0.62.0 (2024-12-16)
* **NEW**: implemented public API **v0.52.1**.
* **NEW**: updated dependencies.
* **NEW**: **kawaii**: added support IPsec features (strongswan managed-app backend).
* **NEW**: **kawaii**: updated metadata scheme.
* **NEW**: **cloud-init**: extended Windows template with NuGet, Paket and PsExec packages installation.
* **BUG**: **cloud-init**: ensure proper Windows admin password setting.
* **BUG**: **kowarp**: fixed various linting issues.

## 0.61.0 (2024-10-17)
* **NEW**: implemented 2-steps user password recovery for security purpose.
* **NEW**: implemented new user session logout and self password reset API calls.
* **NEW**: export user role information in JWT session token (required for Koala Web UI).

## 0.60.1 (2024-10-11)
* **BUG**: retagged to cope with Debian versionning issue.

## 0.60.0 (2024-10-11)
* **NEW**: improved generated emails and add custom theme support.
* **BUG**: fix SDK server-side code generation (issue with objects nested required parameters).
* **BUG**: fixed Debian packages publishing issues.

## 0.60.0-rc1 (2024-10-10)
* **BREAKING CHANGE**: update to new v0.50.0 API, major resources renaming: KGW to Kawaii, KFS to Kylo, KCE to Kompute, NetGW to Kiwi and Pool to StoragePool.
* **BREAKING CHANGE**: updated database schema, requires documents migration.
* **NEW**: implemented database document migration helpers and per-collection schema versioning.
* **NEW**: implemented new **--migrate** command-line flag to perform live MongoDB collections and documents migration.
* **NEW**: restructured the whole source code tree.
* **NEW**: server-side SDK is now directly built-in and auto-generated, instead of using external one, allows for easier development version pinning at engineering stage.
* **NEW**: updated dependencies.
* **BUG**: **cloud-init**: fix public network adapter gateway.

## 0.51.0 (2024-09-17)
* **NEW**: extended **common** library with **DownloadFromURL()** method to efficiently retrieve remote files from HTTP.
* **NEW**: extended **agent** library with ZSTD stream decompression routine.
* **NEW**: extended **agent** library with QCOW2 image support and raw-format conversion.
* **NEW**: extended **ceph** plugin to use new resource download framework and QCOW2 to RAW disk image conversion.
* **NEW**: updated build and runtime dependancies.

## 0.50.6 (2024-09-13)
* **BUG**: fix KGW NAT routing from private LAN traffic (network loop).

## 0.50.5 (2024-09-09)
* **NEW**: add Makefile **tests** directive.
* **NEW**: added new official infographics sources.
* **NEW**: added tests for Konvey agent templating.
* **BUG**: fix test-suite, now passes successfully.
* **BUG**: fix **konvey** traefik configuration templating.

## 0.50.4 (2024-08-29)
* **NEW**: extended Prometheus metrics with instance-based information.
* **BUG**: fix **kgw** nftables + keepalived configuration templating.

## 0.50.3 (2024-08-28)
* **BUG**: **konvey**: ensure valid Traefik configuration settings.
* **BUG**: **konvey**: enforce cross-hosts selection on failover deployments.
* **BUG**: **konvey**: correctly use provided resource name.
* **BUG**: **kgw**: ensure nf_conntrack kernel driver is properly loaded.

## 0.50.2 (2024-08-27)
* **NEW**: upgraded compiler requirement to Go 1.23.
* **NEW**: updated build and runtime dependancies.
* **BUG**: add missing **konvey** resource name in JSON model output, fixes Terraform state inconsistencies.
* **BUG**: tune-in **kgw** NetFilter conntrack settings.

## 0.50.1 (2024-08-09)
* **NEW**: Added support for public API v0.42.
* **NEW**: Added **Konvey** Layer-4 Network Load-Balancer service.
* **NEW**: Restored internal HAR (Highly Available Resource) support to create failover service instances.
* **NEW**: Extended Zone and Regions internal APIs to guess for best-suited computing hosts.
* **NEW**: Add **kontrollers**: a new type of built-in agent for as-a-service instances, replacing Ansible to auto-configure system apps based on instance metadata. Live service update is triggered through WebSocket RPC notification.
* **NEW**: fully get rid of Ansible services post-provisionning.
* **BUG**: Fixed MZR proper project resources references deletion.
* **BUG**: relax **instance** clean-up, prevent from conflicting conditions and ghost DB objects.
* **BUG**: various dependancies upgrades and CVE fixes.

## 0.40.3 (2024-07-25)
* **NEW**: updated build and runtime dependancies.
* **NEW**: instance metadata now also displays underlying virtualization host identity
* **BUG**: instance metadata does not display KGW specific fields (even if null) on non-KGW instances.
* **BUG**: prometheus metrics now correctly reflect storage pool and project costs
* **BUG**: fix KNA PowerDNS error reflection code

## 0.40.2 (2024-07-11)
* **NEW**: KGW now support dynamic updates of firewall and NAT rules (no resource deletion/creation required anymore).
* **BUG**: whitelist KNA PowerDNS recursor zone creation 422 error code
* **BUG**: fix Windows-based instances CloudInit network configuration

## 0.40.1 (2024-06-28)
* **BUG**: fix KNA DNS zone creation test condition.

## 0.40.0.0 (2024-06-13)
* **NEW**: Updated to Kowabunga API v0.41.0.
* **NEW**: KCE instances now supports more than 2 network interfaces.
* **NEW**: Added new multi-zones-resource (MZR) meta-object for as-a-service instances spread over local zones for high-availability.
* **NEW**: KGW now uses MZR resources to provide true cross-zones redundancy, with per-zone private and public gateways and cross-vnet peerings. Fully automated, cross-networks routing, traffic firewalling and forwarding, and public DNAT port-forwarding.
* **NEW**: down-sized KGW hardware requirements.
* **NEW**: KGW now uses dynamic instance metadata information for self-provisionning instead of static cloudinit ones
* **NEW**: cloudinit metadata are now instructed with Kowabunga specifics
* **NEW**: added built-in **kw-meta** Linux utility to retrieve dynamic instance metadata from API.
* **NEW**: extended router header parsing for instance metadata API and improved query response times
* **NEW**: reserve project-specific local gateway IP addresses.
* **NEW**: add new runtime database schema migration and auto-pruner helpers.
* **NET**: add routines to virtual network resources to find the most appropriate subnet, with enough room to host large services.
* **NEW**: KNA agent now bypasses Iris middleware and used direct connection to Iris' PowerDNS APIs.
* **NEW**: project now provides list of VRRP IDs used by -as-a-service instances, not to be reused by end-user.
* **BUG**: instance and KCE's libvirt generated XML now ensures that network and disk interfaces are sorted in the right order and immutable.
* **BUG**: more robust and resilient resources deletion in case of missing objects or cross-references.

## 0.30.0.0 (2024-05-07)
* **NEW**: Stable v0.30 release

## 0.30.0.0-rc4 (2024-04-16)
* **BUG**: fix pool object database value override at each scan period.
* **BUG**: fix instance update and XML generation.
* **BUG**: fix volume BSON parsing.
* **BUG**: implemented API v0.31 changes; project volume creation depends on region, not zone.
* **BUG**: drastically increase http/wsrpc timeout (support for large templates upload)
* **BUG**: add fallback vlan gateway if no kgw in subnet
* **BUG**: fix users assignment in groups.
* **BUG**: move nfs ganesha api backends management from core to KSA agent.
* **BUG**: kgw cloudinit now uses the right public gateway, not an hardcoded one.
* **NEW**: updated dependancies.

## 0.30.0.0-rc3 (2024-04-10)
* **BUG**: fix KCA memory calculation segv on some particular NUMA architectures.
* **BUG**: updated dep to fix HTTP2 CVE

## 0.30.0.0-rc2 (2024-04-05)
* **BUG**: fix some CI/CD build issues
* **BUG**: force user password and API keys generation not to use symbols that might break stupid JSON/YAML parsers.

## 0.30.0.0-rc1 (2024-03-28)
* **BREAK**: Major public API update
    * Migrated to OpenAPI v3.1.
    * Rely on brand new server-side SDK (every routing engine parts have been replaced).
    * Deprecation of **storage pool** types: **local** capability is gone, **ceph** becomes the only supported backend.
    * Update of **template** resources API, Ceph OS volumes are now automatically created from source HTTP(S) URL.

* **BREAK**: Proper Multi-AZ service readiness
    * **Storage Pools** and **NFS** are now part of region, cross-zones, not zone-bounded anymore.
    * **Virtual Networks** and **Subnets** are now part of region, cross-zones, not zone-bounded anymore.
    * **KFS** and **KGW** as-as-sevices resources are now region-global, not zone-bounded anymore.
    * Resources from all zones should be able to use region's global services.

* **BREAK**: Revamped architecture
    * Introduction of n-tier architecture with **Kowabunga Agents**:
        * **KCA**: *Kowabunga Computing Agent*, locally controlling KVM/libvirt hypervisors.
        * **KNA**: *Kowabunga Networking Agent*, locally controlling Iris network gateways.
        * **KSA**: *Kowabunga Storage Agent*, locally controlling Ceph clusters.
    * Kowabunga is now split between the global, Internet-exposed orchestrator, and datacenter-local agent (KCA, KNA, KSA) instances, managing private local resources. Agents connect to orchestrator through secure WebSocket (bypassing all possible private-DC firewall issues) and gets controlled by reverse-RPC calls.
    * Typical usage workflow translates as **Terraform <-> Client SDK <-> API <-> Kowabunga Orchestrator <-> RPC <-> WSS <-> Local Agent**.
    * Both agents and orchestrator auto-detects peer's failure and automatically reconnect in a progressive manner.
    * All direct interaction from orchestrator to libvirt/Ceph/Iris is now delegated to respective agents.
    * Cloud-init ISO images are now stored on Ceph backend (distributed, ready for instances' migration) and do not require a patched version of libvirt anymore.
    * The libvirt XML resource schema generation has been fully refactored.
    * Orchestrator now features a in-memory ultra-efficient database cache (zero-copy, no garbage collection)

* **NEW**: Introduced user management features
    * Kowabunga now features users and groups of users.
    * Projects (and associated underlying resources) do not belong to anyone anymore and can only be created by users with **superAdmin** or **projectAdmin** role.
    * Users belong to one to several groups.
    * Projects are associated with groups.
    * All individuals from project's groups are allowed to access and administrate project's resources.
    * Orchestrator provides robust server-side generated API keys and user passwords, preventing user from using weak credentials
    * User are required to perform a 2-steps account validation upon creation, before being able to consume services.

* **NEW**: Introduced robust authentication mechanisms
    * Kowabunga features 3 ways to consume WebServices:
        * **admin master token** (should never be used, unless for creating first **superAdmin** users)
        * Server-to-server API key authentication
        * JWT-based bearer authentication
    * Orchestrator's HTTP router now features middleware-based API routing layers: log, authentication, authorization, processing
    * Orchestrator's HTTP router uses per API route ACLs checks.

* **NEW**: Miscellaneous features
    * New modular debian packaging, differentiating orchestrator from agents, with multi-architecture support (x86_64, arm64 ...)
    * Enforced Go 1.22 compiler
    * Fixes all known CVEs (to date)
    * Modularized Go packages with dynamic plugin support
    * Support for MacOS targets

* **BUG**: Ensure proper hardware stop at KCE's deletion: ensure proper volume erasure

## 0.10.1.1
* **NEW**: Updates default route to KGW

## 0.10.1.0
* **NEW**: ability to update KGW in a HA manner

## 0.10.0.1
* **BUG**: add missing project's KGW listing API call prototype.

## 0.10.0.0
* **NEW**: implemented project cost retrieval API
* **NEW**: implemented instance remote connection URL retrieval API
* **NEW**: implemented new special /latest/meta-data instance metadata API (AWS-style) to be queried by live instances to retrieve configuration properties
* **NEW**: re-implemented cost management API implementation, every resources now has its own price, flagged in DB.
* **NEW**: major DB queries optimization, fast and furious (should always have been done this way)
* **NEW**: Introduced ansible vars to be injected in cloud init template
* **NEW**: Implemented KGW (**Kowabunga Network Gateway**) object, a network gateway as-a-service, providing Internet inbound and outbound traffic to your project.

## 0.9.0.0
* **NEW**: updated Go compiler requirement to 1.21
* **NEW**: updated Kowabunga API to v0.8.0.
* **NEW**: updated dependancies against known CVEs.
* **NEW**: implemented NFS storage definition and KFS resources (Kowabunga File System, NFS-compatible shares).

## 0.8.0.0
* **NEW**: added support for pre-release binary vulnerability check (**make vuln**)
* **NEW**: extended Spice's virtual machine remote-display configuration to bind on all host interfaces
* **NEW**: add support for Windows-OS virtual machines
* **BREAK**: updated configuration file YAML syntax

## 0.7.8.1
* **BUG**: fix API handler web services

## 0.7.8.0
* **NEW**: expose Kowabunga metrics through native Prometheus format (/metrics endpoint)
* **NEW**: new custom HTTP server implementation, allows for multiple custom endpoint handlers
* **BUG**: cloud-init network config shuld not configure DNS settings if adapter doesn't have any associated IP address.

## 0.7.7.0
* **NEW**: implemented dns_record support from API v0.7.7
* **BUG**: only add instance private IP addresses to the internal DNS record

## 0.7.6.7
* **BUG**: fix possible ssh pubkey misformatting at cloud-init generation

## 0.7.6.6
* **BUG**: fix infinite round up for max score calculation on x86_64 archs

## 0.7.6.5
* **BUG**: extra sanity checks on host

## 0.7.6.4
* **BUG**: fix subnet's reserved range model generation

## 0.7.6.3
* **NEW**: improved host instance election algorithm
* **BUG**: prevent some possible nil pointer de-referencing
* **BUG**: fix multi-hosts spreading, ensuring local pool access is done from the host it belongs to

## 0.7.6.2
* fix deb scripts

## 0.7.6.1
* pre/post tasks on Debian packaging

## 0.7.6
* server-side implementation of Kowabunga API v0.7.6

## 0.1
* initial release
