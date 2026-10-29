{ testers, pkgs }:

testers.nixosTest {
  name = "stashctl";

  nodes.server =
    { ... }:
    {
      imports = [ ../../backend/nix/module.nix ];
      services.stashsphere =
        let
          secretConfig = pkgs.writeText "secret.json" (
            builtins.toJSON {
              invites = {
                enabled = true;
                code = "1234";
              };
              # the test talks plain HTTP to the server
              auth.disableSecureCookies = true;
              # emails go to the journal, where the test reads verification codes
              email.backend = "stdout";
            }
          );
        in
        {
          enable = true;
          settings = {
            database = {
              host = "/run/postgresql";
              password = "foo";
            };
          };
          configFiles = [ "${secretConfig}" ];
          usesLocalPostgresql = true;
        };
      services.postgresql = {
        enable = true;
        ensureDatabases = [ "stashsphere" ];
        ensureUsers = [
          {
            name = "stashsphere";
            ensureDBOwnership = true;
          }
        ];
      };
      environment.systemPackages = [ pkgs.stashctl ];
      environment.variables.STASHCTL_HOME = "/tmp/stashctl";
    };

  testScript = ''
    import base64
    import json
    import struct
    import time
    import zlib

    server_url = "http://127.0.0.1:8081"
    cookies = "/tmp/stashctl/cookies.json"

    start_all()
    server.wait_for_unit("stashsphere.service")
    server.wait_until_succeeds(f"${pkgs.curl}/bin/curl -f {server_url}/api/info")

    # login, refresh and register share a rate limit of a burst of 5
    # requests per IP, refilled by one every 12 seconds, so the test logs
    # in with curl as little as possible

    def api(method, path, body=None):
        """Calls the API with curl as alice, with the access token stashctl stored."""
        stored = json.loads(server.succeed(f"cat {cookies}"))
        access = next(
            cookie["raw_cookie"].split(";")[0] for cookie in stored
            if cookie["raw_cookie"].startswith("stashsphere-access=")
        )
        command = (
            f"${pkgs.curl}/bin/curl -sf -X {method} -H 'cookie: {access}' "
            + f"-H 'content-type: application/json' {server_url}{path}"
        )
        if body is not None:
            command += f" -d '{json.dumps(body)}'"
        out = server.succeed(command)
        return json.loads(out) if out.strip() else None

    def register(user):
        body = json.dumps({
            "name": user,
            "email": f"{user}@example.com",
            "password": "secret",
            "inviteCode": "1234",
        })
        server.succeed(
            "${pkgs.curl}/bin/curl -sf -H 'content-type: application/json' "
            + f"-d '{body}' {server_url}/api/user/register"
        )

    def stashctl_json(command, home="/tmp/stashctl"):
        return json.loads(server.succeed(f"STASHCTL_HOME={home} stashctl --json {command}"))

    def write_png(path, width, height):
        """Writes a red PNG image to path on the server."""
        def chunk(kind, data):
            return (
                struct.pack(">I", len(data)) + kind + data
                + struct.pack(">I", zlib.crc32(kind + data))
            )
        rows = b"".join(b"\x00" + b"\xff\x00\x00" * width for _ in range(height))
        png = (
            b"\x89PNG\r\n\x1a\n"
            + chunk(b"IHDR", struct.pack(">IIBBBBB", width, height, 8, 2, 0, 0, 0))
            + chunk(b"IDAT", zlib.compress(rows))
            + chunk(b"IEND", b"")
        )
        server.succeed(f"echo {base64.b64encode(png).decode()} | base64 -d > {path}")

    def png_width(path):
        """The width in the IHDR chunk of the PNG image at path on the server."""
        return int(server.succeed(f"od -An -tu4 --endian=big -j16 -N4 {path}").strip())

    register("alice")
    register("bob")

    with subtest("commands fail with exit code 3 before login"):
        status, out = server.execute("stashctl --json whoami 2>&1")
        assert status == 3, f"exit code {status}: {out}"
        assert "not logged in" in json.loads(out)["error"]["message"], out
        status, out = server.execute("stashctl --json info 2>&1")
        assert status == 3, f"exit code {status}: {out}"

    with subtest("info needs no login"):
        info = json.loads(server.succeed(f"stashctl --server {server_url} --json info"))
        assert info["inviteRequired"] is True and not info["oidcProviders"], info
        out = server.succeed(f"stashctl --server {server_url} info")
        assert "with an invite code" in out and "login providers" not in out, out
        server.fail(f"test -e {cookies}")

    with subtest("a wrong password is rejected"):
        status, out = server.execute(
            f"echo wrong | stashctl --server {server_url} login --email alice@example.com --password-stdin 2>&1"
        )
        assert status == 3, f"exit code {status}: {out}"
        assert out.strip() == "error: Unauthorized (HTTP 401)", out

    with subtest("login stores the session"):
        server.succeed(
            f"echo secret | stashctl --server {server_url} login --email alice@example.com --password-stdin"
        )
        assert server.succeed(f"stat -c %a {cookies}").strip() == "600"

    with subtest("whoami uses the stored server and session"):
        profile = json.loads(server.succeed("stashctl --json whoami"))
        assert profile["email"] == "alice@example.com", profile
        assert profile["name"] == "alice", profile
        assert "alice@example.com" in server.succeed("stashctl whoami")
        assert server_url in server.succeed("stashctl info")

    with subtest("an expired access token is refreshed"):
        stored = json.loads(server.succeed(f"cat {cookies}"))
        without_access = [
            cookie for cookie in stored
            if not cookie["raw_cookie"].startswith("stashsphere-access=")
        ]
        assert len(without_access) < len(stored), stored
        server.succeed(f"cat > {cookies} <<'EOF'\n{json.dumps(without_access)}\nEOF")
        profile = json.loads(server.succeed("stashctl --json whoami"))
        assert profile["email"] == "alice@example.com", profile
        refreshed = server.succeed(f"cat {cookies}")
        assert "stashsphere-access=" in refreshed, refreshed

    alice_id = stashctl_json("whoami")["id"]
    bob_id = next(user["id"] for user in api("GET", "/api/users") if user["name"] == "bob")

    drill = api("POST", "/api/things", {
        "name": "Cordless drill",
        "description": "18V with two batteries",
        "privateNote": "bought 2024",
        "imagesIds": [],
        "properties": [{"type": "float", "name": "weight", "value": 1.5, "unit": "kg"}],
        "quantity": 2,
        "quantityUnit": "pcs",
        "sharingState": "private",
    })
    workshop = api("POST", "/api/lists", {
        "name": "Workshop",
        "thingIds": [drill["id"]],
        "sharingState": "private",
    })
    share = api("POST", "/api/shares", {"targetUserId": bob_id, "objectId": drill["id"]})

    with subtest("things list and show"):
        things = stashctl_json("things list")
        assert things["totalCount"] == 1, things
        assert things["things"][0]["id"] == drill["id"], things
        assert stashctl_json("things list --search cord")["totalCount"] == 1
        assert stashctl_json("things list --search drill")["totalCount"] == 0
        assert stashctl_json("things list --all")["totalCount"] == 1

        table = server.succeed("stashctl things list")
        assert "Cordless drill" in table and "2 pcs" in table, table
        assert table.endswith("\n1 thing\n"), table

        thing = stashctl_json(f"things show {drill['id']}")
        assert thing["properties"] == [{"type": "float", "name": "weight", "value": 1.5, "unit": "kg"}], thing
        shown = server.succeed(f"stashctl things show {drill['id']}")
        for expected in ["Cordless drill", "Workshop", "bob", "weight:  1.5 kg", "bought 2024"]:
            assert expected in shown, f"{expected!r} missing in:\n{shown}"

    with subtest("properties list and show the schemas"):
        def field(out, label):
            line = next(l for l in out.splitlines() if l.startswith(f"{label}:"))
            return line.split(":", 1)[1].strip()

        collection = stashctl_json("properties list")
        assert "length" in collection["schemas"], collection
        assert collection["aliases"]["len"] == "length", collection
        table = server.succeed("stashctl properties list")
        row = next(l for l in table.splitlines() if l.startswith("length "))
        for expected in ["float", "mm, cm, m, km, in, ft", "l, len", "Length"]:
            assert expected in row, row

        # aliases and any case resolve to the property
        out = server.succeed("stashctl properties show LEN")
        assert field(out, "name") == "length" and field(out, "type") == "float", out
        assert field(out, "aliases") == "l, len", out
        assert stashctl_json("properties show len")["schema"]["title"] == "Length"
        out = server.succeed("stashctl properties show colour")
        assert field(out, "type") == "string" and "aliceblue" in field(out, "values"), out
        assert field(server.succeed("stashctl properties show purchased_at"), "type") == "datetime"
        assert field(server.succeed("stashctl properties show managed"), "type") == "boolean"
        status, out = server.execute("stashctl --json properties show nope 2>&1")
        assert status == 5 and "no property schema named nope" in out, f"exit code {status}: {out}"

    with subtest("missing objects exit with code 5"):
        status, out = server.execute("stashctl --json things show nope 2>&1")
        assert status == 5, f"exit code {status}: {out}"
        assert json.loads(out)["error"]["status"] == 404, out
        # IDs may start with a hyphen and are not taken for flags
        status, out = server.execute("stashctl --json things show -nope 2>&1")
        assert status == 5, f"exit code {status}: {out}"

    with subtest("search finds things and lists"):
        result = stashctl_json("search batteries")
        assert [t["id"] for t in result["things"]] == [drill["id"]], result
        result = stashctl_json("search workshop")
        assert [l["id"] for l in result["lists"]] == [workshop["id"]], result
        assert "No things found." in server.succeed("stashctl search workshop")

    with subtest("lists list and show"):
        lists = stashctl_json("lists list")
        assert lists["totalCount"] == 1, lists
        shown = server.succeed(f"stashctl lists show {workshop['id']}")
        assert "Workshop" in shown and "Cordless drill" in shown, shown

    with subtest("users list and show"):
        users = stashctl_json("users list")
        assert {alice_id, bob_id} <= {u["id"] for u in users}, users
        assert stashctl_json(f"users show {bob_id}")["name"] == "bob"

    with subtest("shares show"):
        shown = stashctl_json(f"shares show {share['id']}")
        assert shown["type"] == "thing", shown
        assert shown["targetUser"]["id"] == bob_id, shown
        assert "shared with:  bob" in server.succeed(f"stashctl shares show {share['id']}")

    with subtest("notifications list for the receiving user"):
        # wait for the login rate limit to allow one more request
        time.sleep(12)
        server.succeed(
            f"echo secret | STASHCTL_HOME=/tmp/bob stashctl --server {server_url} "
            + "login --email bob@example.com --password-stdin"
        )
        notifications = stashctl_json("notifications list --unacknowledged", home="/tmp/bob")
        assert notifications["totalCount"] == 1, notifications
        notification = notifications["notifications"][0]
        assert notification["contentType"] == "THING_SHARED", notification
        assert notification["content"]["thingId"] == drill["id"], notification
        table = server.succeed("STASHCTL_HOME=/tmp/bob stashctl notifications list")
        assert f"thing {drill['id']} shared by user {alice_id}" in table, table

    with subtest("notifications acknowledge"):
        out = server.succeed(
            f"STASHCTL_HOME=/tmp/bob stashctl notifications ack {notification['id']}"
        )
        assert out.strip() == "Acknowledged 1 notification", out
        notifications = stashctl_json("notifications list --unacknowledged", home="/tmp/bob")
        assert notifications["totalCount"] == 0, notifications
        status, out = server.execute("STASHCTL_HOME=/tmp/bob stashctl --json notifications ack nope 2>&1")
        assert status != 0, f"exit code {status}: {out}"

    with subtest("things summary counts own and shared things"):
        summary = stashctl_json("things summary")
        assert summary == {"totalCount": 1, "ownerIds": [alice_id]}, summary
        # bob owns nothing, but sees the drill through its share
        summary = stashctl_json("things summary", home="/tmp/bob")
        assert summary == {"totalCount": 1, "ownerIds": [alice_id]}, summary
        out = server.succeed("STASHCTL_HOME=/tmp/bob stashctl things summary")
        assert out == f"1 thing owned by\n  alice ({alice_id})\n", out

    with subtest("things create"):
        saw = stashctl_json(
            "things create --name 'Hand saw' --quantity 1 --unit pcs "
            + "--property brand=Bahco --property 'length:float=50 cm' "
            + "--property sharp:boolean=true --property bought:datetime=2025-03-01"
        )
        thing = stashctl_json(f"things show {saw['id']}")
        assert thing["name"] == "Hand saw", thing
        assert thing["sharingState"] == "private", thing
        assert thing["properties"] == [
            {"type": "string", "name": "brand", "value": "Bahco"},
            {"type": "float", "name": "length", "value": 50, "unit": "cm"},
            {"type": "boolean", "name": "sharp", "value": True},
            {"type": "datetime", "name": "bought", "value": "2025-03-01T00:00:00Z"},
        ], thing
        out = server.succeed("stashctl things create --name 'Spare blade' --sharing friends")
        assert out.startswith("Created thing Spare blade ("), out
        status, out = server.execute("stashctl things create --name x --property w:float=abc 2>&1")
        assert status == 2, f"exit code {status}: {out}"

    with subtest("things update keeps the fields it does not change"):
        thing = stashctl_json(
            f"things update {drill['id']} --quantity 3 --property 'weight:float=1.7 kg' "
            + "--property color=green"
        )
        assert thing["quantity"] == 3, thing
        assert thing["name"] == "Cordless drill", thing
        assert thing["privateNote"] == "bought 2024", thing
        assert thing["quantityUnit"] == "pcs", thing
        assert thing["properties"] == [
            {"type": "float", "name": "weight", "value": 1.7, "unit": "kg"},
            {"type": "string", "name": "color", "value": "green"},
        ], thing
        thing = stashctl_json(f"things show {drill['id']}")
        assert [l["id"] for l in thing["lists"]] == [workshop["id"]], thing
        assert [s["id"] for s in thing["shares"]] == [share["id"]], thing

        shown = server.succeed(
            f"stashctl things update {saw['id']} --description 'pull saw' --remove-property brand"
        )
        assert "pull saw" in shown and "brand" not in shown, shown
        assert "length:  50 cm" in shown, shown

    with subtest("only the owner can change a thing"):
        status, out = server.execute(
            f"STASHCTL_HOME=/tmp/bob stashctl --json things update {drill['id']} --quantity 9 2>&1"
        )
        assert status == 4, f"exit code {status}: {out}"
        assert stashctl_json(f"things show {drill['id']}")["quantity"] == 3

    with subtest("lists create, update and delete"):
        garden = stashctl_json(f"lists create --name Garden --thing {saw['id']}")
        assert stashctl_json(f"lists show {garden['id']}")["sharingState"] == "private"
        out = server.succeed("stashctl lists create --name Empty")
        assert out.startswith("Created list Empty ("), out

        shed = stashctl_json(
            f"lists update {garden['id']} --name 'Garden shed' --sharing friends "
            + f"--add-thing {drill['id']} --remove-thing {saw['id']}"
        )
        assert shed["name"] == "Garden shed", shed
        assert shed["sharingState"] == "friends", shed
        assert [t["id"] for t in shed["things"]] == [drill["id"]], shed
        shed = stashctl_json(f"lists update {garden['id']} --add-thing {saw['id']}")
        assert shed["name"] == "Garden shed", shed
        assert sorted(t["id"] for t in shed["things"]) == sorted([drill["id"], saw["id"]]), shed

        out = server.succeed(f"stashctl lists delete {garden['id']}")
        assert out.strip() == f"Deleted list {garden['id']}", out
        status, out = server.execute(f"stashctl --json lists show {garden['id']} 2>&1")
        assert status == 5, f"exit code {status}: {out}"
        stashctl_json(f"things show {saw['id']}")

    with subtest("shares create and delete"):
        list_share = stashctl_json(f"shares create {workshop['id']} --user {bob_id}")
        assert list_share["type"] == "list", list_share
        assert list_share["targetUser"]["id"] == bob_id, list_share
        shown = stashctl_json(f"lists show {workshop['id']}", home="/tmp/bob")
        assert shown["name"] == "Workshop", shown

        server.succeed(f"stashctl shares delete {list_share['id']}")
        status, out = server.execute(f"STASHCTL_HOME=/tmp/bob stashctl --json lists show {workshop['id']} 2>&1")
        assert status == 4, f"exit code {status}: {out}"
        status, out = server.execute(f"stashctl --json shares show {list_share['id']} 2>&1")
        assert status == 5, f"exit code {status}: {out}"

    with subtest("images upload and list"):
        write_png("/tmp/wide.png", 40, 20)
        write_png("/tmp/tall.png", 20, 40)
        uploaded = stashctl_json("images upload /tmp/wide.png /tmp/tall.png")
        assert [i["name"] for i in uploaded] == ["wide.png", "tall.png"], uploaded
        wide, tall = uploaded
        assert (wide["width"], wide["height"]) == (40, 20), wide

        images = stashctl_json("images list --unassigned")
        assert images["totalCount"] == 2, images
        table = server.succeed("stashctl images list")
        assert "40x20" in table and wide["hash"] in table, table
        assert table.endswith("\n2 images\n"), table

        status, out = server.execute("echo nope > /tmp/nope.png && stashctl images upload /tmp/nope.png 2>&1")
        assert status == 1, f"exit code {status}: {out}"

    with subtest("things take images"):
        thing = stashctl_json(
            f"things update {drill['id']} --add-image {wide['id']} --add-image {tall['id']}"
        )
        assert [i["id"] for i in thing["images"]] == [wide["id"], tall["id"]], thing
        thing = stashctl_json(f"things update {drill['id']} --remove-image {tall['id']}")
        assert [i["id"] for i in thing["images"]] == [wide["id"]], thing
        assert thing["quantity"] == 3, thing
        shown = server.succeed(f"stashctl things show {drill['id']}")
        assert f"{wide['id']} wide.png 40x20 (hash {wide['hash']})" in shown, shown
        assert [i["id"] for i in stashctl_json("images list --unassigned")["images"]] == [tall["id"]]

        tape = stashctl_json(f"things create --name 'Tape measure' --image {tall['id']}")
        thing = stashctl_json(f"things show {tape['id']}")
        assert [i["id"] for i in thing["images"]] == [tall["id"]], thing

    with subtest("images download"):
        out = server.succeed(f"stashctl images download {wide['hash']} -o /tmp/out.png")
        assert out.strip() == "Saved image to /tmp/out.png", out
        server.succeed("cmp /tmp/wide.png /tmp/out.png")
        server.succeed(f"stashctl images download {wide['hash']} > /tmp/stdout.png")
        server.succeed("cmp /tmp/wide.png /tmp/stdout.png")
        server.succeed(f"stashctl images download {wide['hash']} --width 20 -o /tmp/small.png")
        assert png_width("/tmp/small.png") == 20

        # bob sees the image through the share of the drill
        server.succeed(
            f"STASHCTL_HOME=/tmp/bob stashctl images download {wide['hash']} -o /tmp/bob.png"
        )
        server.succeed("cmp /tmp/wide.png /tmp/bob.png")

        status, out = server.execute("stashctl --json images download nope -o /tmp/missing.png 2>&1")
        assert status == 5, f"exit code {status}: {out}"
        server.fail("test -e /tmp/missing.png")

    with subtest("images rotate"):
        rotated = stashctl_json(f"images rotate {wide['id']} right")
        assert (rotated["width"], rotated["height"]) == (20, 40), rotated
        assert rotated["hash"] != wide["hash"], rotated
        server.succeed(f"stashctl images download {rotated['hash']} -o /tmp/rotated.png")
        assert png_width("/tmp/rotated.png") == 20
        out = server.succeed(f"stashctl images rotate {wide['id']} left")
        assert out.strip() == f"Rotated image wide.png ({wide['id']})", out
        assert [i["width"] for i in stashctl_json(f"things show {drill['id']}")["images"]] == [40]

    with subtest("images delete"):
        status, out = server.execute(f"stashctl images delete {wide['id']} 2>&1")
        assert status == 1, f"exit code {status}: {out}"
        assert "in use" in out, out

        server.succeed(f"stashctl things delete {tape['id']}")
        out = server.succeed(f"stashctl images delete {tall['id']}")
        assert out.strip() == f"Deleted image tall.png ({tall['id']})", out
        assert stashctl_json("images list --unassigned")["totalCount"] == 0

    with subtest("things delete"):
        out = server.succeed(f"stashctl things delete {saw['id']}")
        assert out.strip() == f"Deleted thing {saw['id']}", out
        status, out = server.execute(f"stashctl --json things show {saw['id']} 2>&1")
        assert status == 5, f"exit code {status}: {out}"

    with subtest("friends send, accept and reject requests"):
        # wait for the rate limit to allow registering and logging in carol
        time.sleep(24)
        register("carol")
        server.succeed(
            f"echo secret | STASHCTL_HOME=/tmp/carol stashctl --server {server_url} "
            + "login --email carol@example.com --password-stdin"
        )
        carol_id = stashctl_json("whoami", home="/tmp/carol")["id"]

        to_bob = stashctl_json(f"friends send {bob_id}")
        assert to_bob["state"] == "pending", to_bob
        assert to_bob["receiver"]["id"] == bob_id, to_bob
        status, out = server.execute(f"stashctl friends send {bob_id} 2>&1")
        assert status == 1 and "pending request" in out, f"exit code {status}: {out}"

        requests = stashctl_json("friends requests", home="/tmp/bob")
        assert [r["id"] for r in requests["received"]] == [to_bob["id"]], requests
        table = server.succeed("STASHCTL_HOME=/tmp/bob stashctl friends requests")
        assert to_bob["id"] in table and f"alice ({alice_id})" in table and "pending" in table, table
        table = server.succeed("STASHCTL_HOME=/tmp/bob stashctl notifications list")
        assert f"friend request {to_bob['id']} from user {alice_id}" in table, table

        # only the receiver answers a request
        status, out = server.execute(f"stashctl --json friends accept {to_bob['id']} 2>&1")
        assert status == 4, f"exit code {status}: {out}"
        out = server.succeed(f"STASHCTL_HOME=/tmp/bob stashctl friends accept {to_bob['id']}")
        assert out.strip() == f"Accepted friend request from alice ({alice_id})", out
        status, out = server.execute(f"STASHCTL_HOME=/tmp/bob stashctl friends reject {to_bob['id']} 2>&1")
        assert status == 1 and "not pending" in out, f"exit code {status}: {out}"

        to_carol = stashctl_json(f"friends send {carol_id}")
        rejected = stashctl_json(f"friends reject {to_carol['id']}", home="/tmp/carol")
        assert rejected["state"] == "rejected", rejected
        to_carol = stashctl_json(f"friends send {carol_id}")
        stashctl_json(f"friends accept {to_carol['id']}", home="/tmp/carol")

        friends = stashctl_json("friends list")
        assert sorted(f["friend"]["id"] for f in friends["friendShips"]) == sorted([bob_id, carol_id]), friends
        table = server.succeed("stashctl friends list")
        assert "bob" in table and "carol" in table, table

    with subtest("friends see what is shared with friends"):
        spare = stashctl_json("things list --search spare")["things"][0]
        assert spare["sharingState"] == "friends", spare
        assert stashctl_json("things list --search spare", home="/tmp/carol")["totalCount"] == 1

    with subtest("friends remove and cancel"):
        out = server.succeed(f"stashctl friends remove {carol_id}")
        assert out.strip() == f"Removed friend {carol_id}", out
        friends = stashctl_json("friends list")
        assert [f["friend"]["id"] for f in friends["friendShips"]] == [bob_id], friends
        assert stashctl_json("friends list", home="/tmp/carol")["friendShips"] == []
        assert stashctl_json("things list --search spare", home="/tmp/carol")["totalCount"] == 0
        status, out = server.execute(f"stashctl --json friends remove {carol_id} 2>&1")
        assert status == 5, f"exit code {status}: {out}"

        request = stashctl_json(f"friends send {carol_id}")
        status, out = server.execute(f"STASHCTL_HOME=/tmp/carol stashctl --json friends cancel {request['id']} 2>&1")
        assert status == 4, f"exit code {status}: {out}"
        out = server.succeed(f"stashctl friends cancel {request['id']}")
        assert out.strip() == f"Cancelled friend request {request['id']}", out
        received = stashctl_json("friends requests", home="/tmp/carol")["received"]
        assert request["id"] not in [r["id"] for r in received], received
        assert "No friends yet." in server.succeed("STASHCTL_HOME=/tmp/carol stashctl friends list")

    with subtest("links create, list and show"):
        link = stashctl_json(f"links create {drill['id']}")
        assert link["url"] == f"{server_url}/public/{link['id']}", link
        out = server.succeed(f"stashctl links create {workshop['id']}")
        list_link = out.strip().rsplit("/", 1)[1]
        assert out.strip() == f"{server_url}/public/{list_link}", out
        status, out = server.execute(f"STASHCTL_HOME=/tmp/bob stashctl --json links create {drill['id']} 2>&1")
        assert status == 4, f"exit code {status}: {out}"

        links = stashctl_json("links list")
        assert {l["id"] for l in links} == {link["id"], list_link}, links
        table = server.succeed("stashctl links list")
        assert f"Cordless drill ({drill['id']})" in table and link["url"] in table, table
        assert f"public links:  {link['id']}" in server.succeed(f"stashctl things show {drill['id']}")
        assert f"public links:  {list_link}" in server.succeed(f"stashctl lists show {workshop['id']}")

        # anyone with the token sees the thing, but not its private note
        anonymous = f"STASHCTL_HOME=/tmp/anonymous stashctl --server {server_url}"
        out = server.succeed(f"{anonymous} links show {link['id']}")
        for expected in ["Cordless drill", "owner:        alice", "weight:  1.7 kg"]:
            assert expected in out, out
        assert "bought 2024" not in out, out
        shown = json.loads(server.succeed(f"{anonymous} --json links show {list_link}"))
        assert [t["id"] for t in shown["list"]["things"]] == [drill["id"]], shown
        assert "Cordless drill" in server.succeed(f"{anonymous} links show {list_link}")

    with subtest("links delete"):
        status, out = server.execute(f"STASHCTL_HOME=/tmp/bob stashctl --json links delete {link['id']} 2>&1")
        assert status == 4, f"exit code {status}: {out}"
        out = server.succeed(f"stashctl links delete {link['id']}")
        assert out.strip() == f"Revoked public link {link['id']}", out
        status, out = server.execute(f"{anonymous} --json links show {link['id']} 2>&1")
        assert status == 5, f"exit code {status}: {out}"
        assert [l["id"] for l in stashctl_json("links list")] == [list_link]

    with subtest("profile show and update"):
        out = server.succeed("stashctl profile show")
        assert alice_id in out and "alice@example.com" in out, out
        profile = stashctl_json(
            "profile update --full-name 'Alice Example' --information 'Has tools' --sharing friends"
        )
        assert profile["name"] == "alice", profile
        assert profile["fullName"] == "Alice Example", profile
        assert profile["information"] == "Has tools", profile
        assert profile["defaultSharingState"] == "friends", profile

        write_png("/tmp/me.png", 8, 8)
        me = stashctl_json("images upload /tmp/me.png")[0]
        profile = stashctl_json(f"profile update --image {me['id']}")
        assert profile["image"]["id"] == me["id"], profile
        assert profile["fullName"] == "Alice Example", profile
        status, out = server.execute(f"STASHCTL_HOME=/tmp/bob stashctl --json profile update --image {me['id']} 2>&1")
        assert status == 4, f"exit code {status}: {out}"
        out = server.succeed("stashctl profile update --remove-image --name alice")
        assert "image:" not in out and "default sharing:  friends" in out, out
        assert stashctl_json("profile show")["image"] is None

    with subtest("profile verifies the email address"):
        out = server.succeed("stashctl profile send-verification")
        assert out.strip() == "Sent a verification code to alice@example.com", out
        code = server.succeed(
            "journalctl -u stashsphere.service --no-pager "
            + "| grep -o 'Your verification code is: [0-9]*' | tail -n 1 | grep -o '[0-9]*$'"
        ).strip()
        assert len(code) == 8, code
        # verifying shares the rate limit of login, wait for two requests
        time.sleep(24)
        status, out = server.execute("stashctl --json profile verify-email 00000000 2>&1")
        assert status == 1, f"exit code {status}: {out}"
        out = server.succeed(f"stashctl profile verify-email {code}")
        assert out.strip() == "Verified your email address", out
        assert "alice@example.com (verified)" in server.succeed("stashctl profile show")

    with subtest("profile password"):
        carol = "STASHCTL_HOME=/tmp/carol stashctl"
        status, out = server.execute(f"printf 'wrong\\nnewsecret\\n' | {carol} --json profile password --password-stdin 2>&1")
        assert status == 1 and "Incorrect old password" in out, f"exit code {status}: {out}"
        status, out = server.execute(f"echo secret | {carol} --json profile password --password-stdin 2>&1")
        assert status == 1 and "two lines" in out, f"exit code {status}: {out}"
        out = server.succeed(f"printf 'secret\\nnewsecret\\n' | {carol} profile password --password-stdin")
        assert out.strip() == "Changed your password", out

    with subtest("profile delete-account and cancel-deletion"):
        # scheduling a deletion checks the password, the old one is gone
        status, out = server.execute(f"echo secret | {carol} --json profile delete-account --password-stdin 2>&1")
        assert status == 1 and "Incorrect password" in out, f"exit code {status}: {out}"
        out = server.succeed(f"echo newsecret | {carol} profile delete-account --password-stdin")
        assert out.startswith("Your account will be deleted at "), out
        assert stashctl_json("profile show", home="/tmp/carol")["purgeAt"] is not None
        assert "deletion:" in server.succeed(f"{carol} profile show")
        out = server.succeed(f"{carol} profile cancel-deletion")
        assert out.strip() == "Cancelled the deletion of your account", out
        assert stashctl_json("profile show", home="/tmp/carol")["purgeAt"] is None
        status, out = server.execute(f"{carol} --json profile cancel-deletion 2>&1")
        assert status == 5, f"exit code {status}: {out}"

    with subtest("export start, status and download"):
        status, out = server.execute("stashctl --json export status 2>&1")
        assert status == 5, f"exit code {status}: {out}"
        status, out = server.execute("stashctl --json export download 2>&1")
        assert status == 5, f"exit code {status}: {out}"
        export = stashctl_json("export start")
        assert export["status"] in ("pending", "processing"), export
        status, out = server.execute("stashctl --json export start 2>&1")
        assert status == 1 and "already in progress" in out, f"exit code {status}: {out}"
        # the export worker picks up new exports once a minute
        done = stashctl_json("export status --wait")
        assert done["id"] == export["id"] and done["status"] == "done", done
        assert done["expiresAt"] is not None, done

        out = server.succeed("mkdir -p /tmp/exports && cd /tmp/exports && stashctl export download")
        archive = "/tmp/exports/" + out.strip().removeprefix("Saved export to ")
        assert archive.startswith("/tmp/exports/stashsphere-export-") and archive.endswith(".zip"), out
        assert "collection.json" in server.succeed(f"${pkgs.unzip}/bin/unzip -l {archive}")
        server.succeed(f"stashctl export download -o - > /tmp/stdout.zip && cmp /tmp/stdout.zip {archive}")

    with subtest("import start and status"):
        status, out = server.execute(f"{carol} --json import status 2>&1")
        assert status == 5, f"exit code {status}: {out}"
        # the upload is accepted and fails in the import worker
        status, out = server.execute(f"{carol} --json import start /tmp/me.png --wait 2>&1")
        assert status == 1 and "invalid zip" in out, f"exit code {status}: {out}"
        assert stashctl_json("import status", home="/tmp/carol")["status"] == "error"

        imported = stashctl_json(f"import start {archive} --wait", home="/tmp/carol")
        assert imported["status"] == "done" and imported["thingsImported"] > 0, imported
        out = server.succeed(f"{carol} import status")
        assert "done" in out and f"things:     {imported['thingsImported']}" in out, out
        drills = stashctl_json("things list --search cordless", home="/tmp/carol")["things"]
        assert [t["owner"]["id"] for t in drills] == [carol_id], drills

    with subtest("cart add, show, remove and clear"):
        bob = "STASHCTL_HOME=/tmp/bob stashctl"
        assert stashctl_json("cart show", home="/tmp/bob")["entries"] == []
        assert server.succeed(f"{bob} cart show").strip() == "Your cart is empty."

        # bob sees the drill through its share, and adding twice keeps one entry
        cart = stashctl_json(f"cart add {drill['id']} {drill['id']}", home="/tmp/bob")
        assert [(e["thingId"], e["ownerId"]) for e in cart["entries"]] == [(drill["id"], alice_id)], cart
        cart = stashctl_json(f"cart add {drill['id']}", home="/tmp/bob")
        assert len(cart["entries"]) == 1, cart
        out = server.succeed(f"{bob} cart show")
        assert f"Owned by alice ({alice_id}):" in out and "Cordless drill" in out, out

        # bob does not see the drill carol imported
        status, out = server.execute(f"{bob} --json cart add {drills[0]['id']} 2>&1")
        assert status == 4, f"exit code {status}: {out}"
        status, out = server.execute(f"{bob} --json cart add nope 2>&1")
        assert status == 5, f"exit code {status}: {out}"
        assert len(stashctl_json("cart show", home="/tmp/bob")["entries"]) == 1

        assert stashctl_json(f"cart remove {drill['id']}", home="/tmp/bob")["entries"] == []
        stashctl_json(f"cart add {drill['id']}", home="/tmp/bob")
        assert server.succeed(f"{bob} cart clear").strip() == "Your cart is empty."
        assert stashctl_json("cart show", home="/tmp/bob")["entries"] == []

    with subtest("STASHCTL_OUTPUT selects JSON, compact when not on a terminal"):
        out = server.succeed("STASHCTL_OUTPUT=json stashctl whoami")
        assert out.count("\n") == 1, out
        assert json.loads(out)["email"] == "alice@example.com", out

    with subtest("logout removes the session"):
        server.succeed("stashctl logout")
        server.fail(f"test -e {cookies}")
        status, out = server.execute("stashctl --json whoami 2>&1")
        assert status == 3, f"exit code {status}: {out}"
  '';
}
