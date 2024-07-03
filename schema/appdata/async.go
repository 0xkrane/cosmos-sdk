package appdata

func AsyncListener(listener Listener, bufferSize int, commitChan chan<- error, doneChan <-chan struct{}) Listener {
	packetChan := make(chan Packet, bufferSize)
	res := Listener{}

	go func() {
		var err error
		for {
			select {
			case packet := <-packetChan:
				if err != nil {
					// if we have an error, don't process any more packets
					// and return the error and finish when it's time to commit
					if _, ok := packet.(CommitData); ok {
						commitChan <- err
						return
					}
				} else {
					// process the packet
					err = listener.SendPacket(packet)
					// if it's a commit
					if _, ok := packet.(CommitData); ok {
						commitChan <- err
						if err != nil {
							return
						}
					}
				}

			case <-doneChan:
				return
			}
		}
	}()

	if listener.InitializeModuleData != nil {
		res.InitializeModuleData = func(data ModuleInitializationData) error {
			packetChan <- data
			return nil
		}
	}

	if listener.StartBlock != nil {
		res.StartBlock = func(data StartBlockData) error {
			packetChan <- data
			return nil
		}
	}

	if listener.OnTx != nil {
		res.OnTx = func(data TxData) error {
			packetChan <- data
			return nil
		}
	}

	if listener.OnEvent != nil {
		res.OnEvent = func(data EventData) error {
			packetChan <- data
			return nil
		}
	}

	if listener.OnKVPair != nil {
		res.OnKVPair = func(data KVPairData) error {
			packetChan <- data
			return nil
		}
	}

	if listener.OnObjectUpdate != nil {
		res.OnObjectUpdate = func(data ObjectUpdateData) error {
			packetChan <- data
			return nil
		}
	}

	if listener.Commit != nil {
		res.Commit = func(data CommitData) error {
			packetChan <- data
			return nil
		}
	}

	return res
}

func AsyncListenerMux(listeners []Listener, bufferSize int, doneChan <-chan struct{}) Listener {
	asyncListeners := make([]Listener, len(listeners))
	commitChans := make([]chan error, len(listeners))
	for i, l := range listeners {
		commitChan := make(chan error)
		commitChans[i] = commitChan
		asyncListeners[i] = AsyncListener(l, bufferSize, commitChan, doneChan)
	}
	mux := ListenerMux(asyncListeners...)
	muxCommit := mux.Commit
	mux.Commit = func(data CommitData) error {
		err := muxCommit(data)
		if err != nil {
			return err
		}

		for _, commitChan := range commitChans {
			err := <-commitChan
			if err != nil {
				return err
			}
		}
		return nil
	}

	return mux
}
