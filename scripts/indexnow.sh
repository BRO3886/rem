#!/usr/bin/env sh
# Submit this site's live URLs to IndexNow (Bing/Microsoft + Yandex/Naver/Seznam/Yep).
# Run after a deploy whose content changed. The key file must be live at the site
# root before submissions will be accepted.
set -eu

HOST="rem.sidv.dev"
KEY="785f0399d4f64b1d1775006ca113f39f"
KEYLOC="https://${HOST}/${KEY}.txt"
ENDPOINT="https://api.indexnow.org/indexnow"

urls=$(curl -fsS "https://${HOST}/sitemap.xml" | grep -o '<loc>[^<]*</loc>' | sed 's/<loc>//; s|</loc>||')
[ -n "$urls" ] || { echo "no URLs found in sitemap"; exit 1; }

count=0; fail=0
for u in $urls; do
  code=$(curl -s -o /dev/null -w '%{http_code}' "${ENDPOINT}?url=${u}&key=${KEY}&keyLocation=${KEYLOC}")
  case "$code" in
    200|202) count=$((count+1)) ;;
    *) echo "FAILED ($code): $u"; fail=1 ;;
  esac
done

echo "Submitted ${count} URL(s) to IndexNow"
[ "$fail" -eq 0 ] || exit 1
