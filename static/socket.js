//@ts-check

export default class Socket {
  #socket;

  /**
   * @param {string} roomCode
   * @param {string} playerID
   */
  constructor(roomCode, playerID) {
    this.#socket = new WebSocket(
      location.origin.replace(/^http/, "ws") +
        `/room/${roomCode}/player/${playerID}`
    );
  }

  /**
   *
   * @param {(data: any) => void} callback
   * @returns {() => void}
   */
  onUpdate(callback) {
    const handler = (event) => {
      callback(JSON.parse(event.data));
    };
    this.#socket.addEventListener("message", handler);

    return () => {
      this.#socket.removeEventListener("message", handler);
    };
  }

  /**
   * @param {number} change
   */
  updateLife(change) {
    this.#send("change", change);
  }

  /**
   * @param {string} name
   */
  setName(name) {
    this.#send("set-name", name);
  }

  /**
   *
   * @param {string} type
   * @param {any} data
   */
  #send(type, data) {
    this.#socket.send(
      JSON.stringify({
        type: type,
        data: data,
      })
    );
  }
}
