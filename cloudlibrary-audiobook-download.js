(async () => {
  const state = window.__remixContext.state.loaderData["routes/listen.$id"];

  const bookID = state.fulfillmentId;
  const sessionKey = state.session_key;
  const licenseId = state.licenseId;
  const title = state.title.replace(/[^a-zA-Z0-9]/g, "_").replace(/_+/g, "_")

  const playlist = await fetch(
    `https://api.findawayworld.com/v4/audiobooks/${bookID}/playlists`,
    {
      headers: {
        "session-key": sessionKey,
      },
      referrer: "https://audio.yourcloudlibrary.com/",
      body: JSON.stringify({
        license_id: licenseId,
      }),
      method: "POST",
      mode: "cors",
      credentials: "omit",
    },
  ).then(r => r.json());

  const dirHandle = await window.showDirectoryPicker({
    mode: "readwrite",
    startIn: "documents"
  });

  let i = 0;
  for (const track of playlist.playlist) {
    const resp = await fetch(track.url, {
      "referrer": "https://audio.yourcloudlibrary.com/",
      "method": "GET",
      "mode": "cors",
      "credentials": "omit"
    });
    i++;
    const fileName = `${title}_${String(i).padStart(3, '0')}.mp3`
    console.log(`${fileName} ${i}/${playlist.playlist.length}`);
    const fileHandle = await dirHandle.getFileHandle(fileName, {create: true})
    const w = await fileHandle.createWritable()
    await resp.body.pipeTo(w)
  }

  console.log('Done');
})();

/*
window.__remixContext = {
  future: {
    v3_fetcherPersist: false,
    v3_relativeSplatPath: false,
    v3_throwAbortReason: false,
    v3_routeConfig: false,
    v3_singleFetch: false,
    v3_lazyRouteDiscovery: false,
    unstable_optimizeDeps: false,
  },
  state: {
    loaderData: {
      root: {},
      "routes/listen.$id": {
        auth_token:
          "eyJhbGciOiJSUzI1NiIsImtpZCI6IjgzQUY2NkZGRTY4QjJDODI0NEZBMzlFNEMyOUYyMDk1MURDMEREMzAiLCJ0eXAiOiJKV1QifQ.eyJjbGllbnRfb3JnX2tleXMiOiJtcy1kYzE3ZTBiNTNiOGU0N2Q3YjQyNzA4Nzc4NzMzZGU3NixzYS0yMkNERUFERS1GRUUwLUU3MTEtMkE5NC0wMDI0RTg4MEEyQjcsY2wtaHZlbWYscnQtQnMxUFZmc0Y3VXBMeHhYIiwibXNzLXBhdHJvbklkIjoiYTNkYWM3NDgtNDYyNC00MjBmLTg5MjgtYzlmYWE4ZDgyYTQ3IiwibXNzLXBhdHJvbkxpYnJhcnlJZCI6ImRjMTdlMGI1M2I4ZTQ3ZDdiNDI3MDg3Nzg3MzNkZTc2IiwibXNzLXJlYWt0b3JUb2tlbiI6ImtpVHZUb2xHNDNrLUNQSEtOdlcqLUxnWUVqKktObHlZWG9kRFFhakJyTC1AIiwic2NvcGUiOiJtc3MtcGF0cm9uYXBpLWFsbCBtc3Mtc2Vuc2l0aXZlbGlicmFyeWNvbmZpZy12aWV3IiwibmJmIjoxNzgxMjc1NTYyLCJleHAiOjE3ODEzNjE5NjIsImlzcyI6Imh0dHBzOi8vbXNzYXV0aC55b3VyY2xvdWRsaWJyYXJ5LmNvbSIsImF1ZCI6Im1zcy1wYXRyb25BcHBzIn0.keg1G1gRBH-odc7YymQuIKsYBxMl-Hb85csqkZnTP74Vgsr-uYJz4bXABTcPyfheKgbLzG2l2QeYXILrK4ftK8AmXakaAmf46KNdFK6_wRqCJTenhwgxGyWKAkzagH1OW77Brp1kR8sYcwUS4wykIdtO0AqfBBIq1uURXc1oe_laFoGQsZbfTbZ5XvgevjJ_ifjrrL55vSk7ZRToe3ZPVUnxlgLsDZ7_hv7dEVrrRaR1WwylyHlorSAXVxjXL_VVt-CeNScKWo32BNIeoRioA7KHAZwUSsaoPEoHQc1E2P_P0BvkTKRx-rIR1pCloiR3w76AlR1meU6sBaoACQl7BA",
        base_url: "https://audio.yourcloudlibrary.com",
        initLastReadingPosition: { offset: 6, part: 1, chapter: 2 },
        initBookmarkPositions: [],
        patron: "kiTvTolG43k-CPHKNvW*-LgYEj*KNlyYXodDQajBrL-@",
        library: "dc17e0b53b8e47d7b42708778733de76",
        webpatron: "https://ebook.yourcloudlibrary.com/library/guelphpl",
        debounceTimeoutLoaded: 2000,
        debounceTimeoutLRP: 3000,
        debounceTimeoutBookmarks: 1000,
        account: "3M-hvemf",
        session_key: "313260a0-c626-4a03-873f-504e6ebc9988",
        audiobook: {
          format: "MP3",
          items: [
            {
              title: "Part 1 Chapter 1",
              sequence: 1,
              part: 1,
              duration: 24502,
            },
            {
              title: "Part 1 Chapter 2",
              sequence: 2,
              part: 1,
              duration: 428355,
            },
            {
              title: "Part 1 Chapter 3",
              sequence: 3,
              part: 1,
              duration: 862511,
            },
            {
              title: "Part 1 Chapter 4",
              sequence: 4,
              part: 1,
              duration: 1059761,
            },
            {
              title: "Part 1 Chapter 5",
              sequence: 5,
              part: 1,
              duration: 799738,
            },
            {
              title: "Part 1 Chapter 6",
              sequence: 6,
              part: 1,
              duration: 484675,
            },
            {
              title: "Part 1 Chapter 7",
              sequence: 7,
              part: 1,
              duration: 638406,
            },
            {
              title: "Part 1 Chapter 8",
              sequence: 8,
              part: 1,
              duration: 786442,
            },
            {
              title: "Part 1 Chapter 9",
              sequence: 9,
              part: 1,
              duration: 701962,
            },
            {
              title: "Part 1 Chapter 10",
              sequence: 10,
              part: 1,
              duration: 532271,
            },
            {
              title: "Part 1 Chapter 11",
              sequence: 11,
              part: 1,
              duration: 457273,
            },
            {
              title: "Part 1 Chapter 12",
              sequence: 12,
              part: 1,
              duration: 483239,
            },
            {
              title: "Part 1 Chapter 13",
              sequence: 13,
              part: 1,
              duration: 756218,
            },
            {
              title: "Part 1 Chapter 14",
              sequence: 14,
              part: 1,
              duration: 484963,
            },
            {
              title: "Part 1 Chapter 15",
              sequence: 15,
              part: 1,
              duration: 720457,
            },
            {
              title: "Part 1 Chapter 16",
              sequence: 16,
              part: 1,
              duration: 429087,
            },
            {
              title: "Part 1 Chapter 17",
              sequence: 17,
              part: 1,
              duration: 748538,
            },
            {
              title: "Part 1 Chapter 18",
              sequence: 18,
              part: 1,
              duration: 538827,
            },
            {
              title: "Part 1 Chapter 19",
              sequence: 19,
              part: 1,
              duration: 601469,
            },
            {
              title: "Part 1 Chapter 20",
              sequence: 20,
              part: 1,
              duration: 625528,
            },
            {
              title: "Part 1 Chapter 21",
              sequence: 21,
              part: 1,
              duration: 434207,
            },
            {
              title: "Part 1 Chapter 22",
              sequence: 22,
              part: 1,
              duration: 867657,
            },
            {
              title: "Part 1 Chapter 23",
              sequence: 23,
              part: 1,
              duration: 633861,
            },
            {
              title: "Part 1 Chapter 24",
              sequence: 24,
              part: 1,
              duration: 723147,
            },
            {
              title: "Part 1 Chapter 25",
              sequence: 25,
              part: 1,
              duration: 846393,
            },
            {
              title: "Part 1 Chapter 26",
              sequence: 26,
              part: 1,
              duration: 844643,
            },
            {
              title: "Part 1 Chapter 27",
              sequence: 27,
              part: 1,
              duration: 596976,
            },
            {
              title: "Part 1 Chapter 28",
              sequence: 28,
              part: 1,
              duration: 703608,
            },
            {
              title: "Part 1 Chapter 29",
              sequence: 29,
              part: 1,
              duration: 685400,
            },
            {
              title: "Part 1 Chapter 30",
              sequence: 30,
              part: 1,
              duration: 795768,
            },
            {
              title: "Part 1 Chapter 31",
              sequence: 31,
              part: 1,
              duration: 1108218,
            },
            {
              title: "Part 1 Chapter 32",
              sequence: 32,
              part: 1,
              duration: 534517,
            },
            {
              title: "Part 1 Chapter 33",
              sequence: 33,
              part: 1,
              duration: 618422,
            },
            {
              title: "Part 1 Chapter 34",
              sequence: 34,
              part: 1,
              duration: 803761,
            },
            {
              title: "Part 1 Chapter 35",
              sequence: 35,
              part: 1,
              duration: 1182798,
            },
            {
              title: "Part 1 Chapter 36",
              sequence: 36,
              part: 1,
              duration: 633704,
            },
            {
              title: "Part 1 Chapter 37",
              sequence: 37,
              part: 1,
              duration: 515944,
            },
            {
              title: "Part 1 Chapter 38",
              sequence: 38,
              part: 1,
              duration: 841848,
            },
            {
              title: "Part 1 Chapter 39",
              sequence: 39,
              part: 1,
              duration: 549929,
            },
            {
              title: "Part 1 Chapter 40",
              sequence: 40,
              part: 1,
              duration: 440737,
            },
            {
              title: "Part 1 Chapter 41",
              sequence: 41,
              part: 1,
              duration: 606772,
            },
            {
              title: "Part 1 Chapter 42",
              sequence: 42,
              part: 1,
              duration: 661289,
            },
            {
              title: "Part 1 Chapter 43",
              sequence: 43,
              part: 1,
              duration: 502413,
            },
            {
              title: "Part 1 Chapter 44",
              sequence: 44,
              part: 1,
              duration: 1005871,
            },
            {
              title: "Part 1 Chapter 45",
              sequence: 45,
              part: 1,
              duration: 640757,
            },
            {
              title: "Part 1 Chapter 46",
              sequence: 46,
              part: 1,
              duration: 298736,
            },
            {
              title: "Part 1 Chapter 47",
              sequence: 47,
              part: 1,
              duration: 31373,
            },
          ],
          sessionKey: "313260a0-c626-4a03-873f-504e6ebc9988",
          accountId: "3M-hvemf",
          checkoutId: "6a2c018aabef987590d05118",
          fulfillmentId: "5179147",
          licenseId: "6932041821a4a73b8c03530d",
        },
        fulfillmentId: "5179147",
        licenseId: "6932041821a4a73b8c03530d",
        isbn: "9781250402950",
        title: "Brigands \u0026 Breadknives",
        backgroundcolor: "#2d2e2d",
      },
    },
    actionData: null,
    errors: null,
  },
};
*/
